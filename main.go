package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/lxn/win"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const btnTextStartCount = "开始数胎动"
const btnTextCountFetalMovement = "胎动了~点一下"
const btnTextStartCountAgain = "重新数胎动"

var (
	mutexCounting sync.Mutex
	counting      bool

	mutexHasEffectiveCount sync.Mutex
	hasEffectiveCount      bool
	hasSetEffectiveCount   bool
	buttonEnabled          = true
	stopCountdown          = make(chan bool)
)

type MyWindow struct {
	*walk.MainWindow
	hWnd                    win.HWND
	PushBtnStart            *walk.PushButton
	PushBtnCancel           *walk.PushButton
	TextLabelCountdown      *walk.TextLabel
	TextLabelActualCount    *walk.TextLabel
	TextLabelEffectiveCount *walk.TextLabel
	ni                      *walk.NotifyIcon
}

func (mw *MyWindow) Start() {
	str := mw.PushBtnStart.Text()

	log.Println(str)
	if str == btnTextStartCount && str == btnTextStartCountAgain {
		mw.PushBtnStart.SetText(btnTextCountFetalMovement)
		mw.PushBtnCancel.SetEnabled(!mw.PushBtnCancel.Enabled())
	}
	if str == btnTextCountFetalMovement {
		actualCount, _ := strconv.Atoi(mw.TextLabelActualCount.Text())
		actualCount++
		mw.TextLabelActualCount.SetText(strconv.Itoa(actualCount))

		if !hasEffectiveCount {
			mutexHasEffectiveCount.Lock()
			hasEffectiveCount = true
			mutexHasEffectiveCount.Unlock()
		}
	}

	// 调用一次倒计时函数
	if !isStartCountdown() {
		mw.startCountdown()
	}

	if buttonEnabled && str == btnTextCountFetalMovement {
		buttonEnabled = false
		mw.PushBtnStart.SetEnabled(false)
		time.AfterFunc(time.Second, func() {
			buttonEnabled = true
			mw.PushBtnStart.SetEnabled(true)
		})
	}
}

func (mw *MyWindow) Cancel() {
	if walk.MsgBox(mw, "确认", "您确定要执行此操作吗？", walk.MsgBoxIconQuestion|walk.MsgBoxYesNo) == walk.DlgCmdNo {
		return
	} else {
		log.Println("点击取消按钮处理")
		str := mw.PushBtnStart.Text()
		if str == btnTextCountFetalMovement {
			mw.PushBtnStart.SetText(btnTextStartCount)
		}
		mw.PushBtnCancel.SetEnabled(!mw.PushBtnCancel.Enabled())

		mw.TextLabelCountdown.SetText("60:00")
		mw.TextLabelActualCount.SetText("0")
		mw.TextLabelEffectiveCount.SetText("0")
		stopCountdown <- true
	}
}

func (mw *MyWindow) removeStyle(style int32) {
	currStyle := win.GetWindowLong(mw.hWnd, win.GWL_STYLE)
	win.SetWindowLong(mw.hWnd, win.GWL_STYLE, currStyle&style)
}

func (mw *MyWindow) AddNotifyIcon() {
	var err error
	mw.ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}

	icon, err := walk.Resources.Image("img/favicon.ico")
	if err != nil {
		log.Fatal(err)
	}
	mw.SetIcon(icon)
	mw.ni.SetIcon(icon)
	mw.ni.SetVisible(true)

	mw.ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			mw.Show()
			win.ShowWindow(mw.Handle(), win.SW_RESTORE)
		}
	})
}

func isStartCountdown() bool {
	mutexCounting.Lock()
	defer mutexCounting.Unlock()
	return counting
}

func (mw *MyWindow) startCountdown() {

	mutexCounting.Lock()
	counting = true
	mutexCounting.Unlock()

	duration := 3600 // 倒计时持续时间，单位为秒

	go func() {
		for i := duration; i > 0; i-- {
			select {
			case <-stopCountdown:
				log.Println("取消倒计时！")
				mutexCounting.Lock()
				counting = false
				mutexCounting.Unlock()
				return
			default:
				minutes := i / 60
				seconds := i % 60
				mw.TextLabelCountdown.SetText(fmt.Sprintf("%02d:%02d", minutes, seconds))

				if i%300 == 0 {
					mw.calcEffectiveCount()
				}
				time.Sleep(time.Second)
			}
		}

		// 显示倒计时结束信息
		mw.TextLabelCountdown.SetText("倒计时结束!")
		mw.PushBtnStart.SetText(btnTextStartCountAgain)

		mutexCounting.Lock()
		counting = false
		mutexCounting.Unlock()
	}()
}

func (mw *MyWindow) setEffectiveCount() {
	mutexHasEffectiveCount.Lock()
	if hasEffectiveCount && !hasSetEffectiveCount {
		effectiveCount, _ := strconv.Atoi(mw.TextLabelEffectiveCount.Text())
		effectiveCount++
		mw.TextLabelEffectiveCount.SetText(strconv.Itoa(effectiveCount))
		hasSetEffectiveCount = true
	}
	mutexHasEffectiveCount.Unlock()
}

// 有效胎动次数为：5分钟内有胎动算1次胎动
func (mw *MyWindow) calcEffectiveCount() {
	log.Println("计算有效胎动次数。。。")
	hasSetEffectiveCount = false

	duration := 100 // 计时5分钟

	go func() {
		for i := duration; i > 0; i-- {
			select {
			case <-stopCountdown:
				log.Println("取消有效胎动次数倒计时！")
				mutexHasEffectiveCount.Lock()
				hasEffectiveCount = false
				hasSetEffectiveCount = false
				mutexHasEffectiveCount.Unlock()
				return
			default:
				mw.setEffectiveCount()
				time.Sleep(3 * time.Second)
			}
		}

		log.Println("有效胎动次数倒计时退出！")
		mutexHasEffectiveCount.Lock()
		hasEffectiveCount = false
		hasSetEffectiveCount = false
		mutexHasEffectiveCount.Unlock()
	}()
}

func main() {
	mw := new(MyWindow)
	if err := (MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "胎动计数器",
		Size:     Size{550, 380},
		Layout:   VBox{MarginsZero: true},
		OnSizeChanged: func() {
			if win.IsIconic(mw.Handle()) {
				mw.Hide()
				mw.ni.SetVisible(true)
			}
		},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{Text: "计时"},
							TextLabel{AssignTo: &mw.TextLabelCountdown, Text: "60:00"},
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{Text: "实际胎动次数"},
							TextLabel{AssignTo: &mw.TextLabelActualCount, Text: "0"},
						},
					},
					Composite{
						Layout: VBox{},
						Children: []Widget{
							Label{Text: "有效胎动次数"},
							TextLabel{AssignTo: &mw.TextLabelEffectiveCount, Text: "0"},
						},
					},
				},
			},
			// HSplitter{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						MaxSize:   Size{Width: 100, Height: 30},
						AssignTo:  &mw.PushBtnCancel,
						Text:      "取消",
						OnClicked: mw.Cancel,
						Enabled:   false,
					},
					PushButton{
						MaxSize:   Size{Width: 100, Height: 30},
						AssignTo:  &mw.PushBtnStart,
						Text:      btnTextStartCount,
						OnClicked: mw.Start,
					},
				},
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}
	mw.hWnd = mw.Handle()
	mw.AddNotifyIcon()

	// 禁止最小化、最大化，禁止修改窗口大小
	mw.removeStyle(^win.WS_MINIMIZEBOX)
	mw.removeStyle(^win.WS_MAXIMIZEBOX)
	mw.removeStyle(^win.WS_SIZEBOX)

	// // 获取当前屏幕信息
	// app := walk.App()
	// screen := app.Screen()
	// screenWidth := screen.WorkArea().Width
	// screenHeight := screen.WorkArea().Height

	// // 计算窗口的居中位置
	// windowWidth := int(mw.Width())
	// windowHeight := int(mw.Height())
	// x := (screenWidth - windowWidth) / 2
	// y := (screenHeight - windowHeight) / 2

	// // 设置窗口位置为居中
	// mw.SetX(x)
	// mw.SetY(y)

	mw.Run()
}
