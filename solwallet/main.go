package main

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"log"
	"os"
	"time"
)

type Wallet struct {
	Index    int
	Mnemonic string
	Private  string
	Address  string
}

type WalletModel struct {
	walk.TableModelBase
	items []Wallet
}

func (m *WalletModel) RowCount() int {
	return len(m.items)
}

func (m *WalletModel) Value(row, col int) interface{} {
	item := m.items[row]
	switch col {
	case 0:
		return item.Index
	case 1:
		return "***"
	case 2:
		return "***"
	case 3:
		return item.Address
	}
	return ""
}

func main() {
	logFile, _ := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	log.SetOutput(logFile)
	log.Println("程序开始启动")

	var mw *walk.MainWindow
	var numberEdit *walk.NumberEdit
	var progressBar *walk.ProgressBar
	var tableView *walk.TableView
	var model *WalletModel

	model = new(WalletModel)
	model.items = make([]Wallet, 0)

	if _, err := (MainWindow{
		AssignTo: &mw,
		Title:    "Solana钱包生成器",
		MinSize:  Size{1200, 800},
		Layout:   VBox{},
		Children: []Widget{
			GroupBox{
				Title:  "设置",
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text:    "生成数量:",
						MinSize: Size{60, 0},
					},
					NumberEdit{
						AssignTo: &numberEdit,
						Value:    1.0,
						MinValue: 1.0,
						MaxValue: 1000.0,
						Decimals: 0,
						MinSize:  Size{60, 0},
					},
					HSpacer{},
					PushButton{
						Text:    "生成钱包",
						MinSize: Size{100, 30},
						OnClicked: func() {
							if btn, ok := mw.Children().At(0).(*walk.GroupBox).Children().At(3).(*walk.PushButton); ok {
								btn.SetEnabled(false)
								defer btn.SetEnabled(true)
							}

							num := int(numberEdit.Value())
							model.items = make([]Wallet, 0, num)
							model.PublishRowsReset()

							go func() {
								allWallets := make([]Wallet, 0, num)
								currentTime := time.Now().Format("20060102150405")
								fileName := fmt.Sprintf("SOL生成%d个_%s.txt", num, currentTime)
								file, err := os.Create(fileName)
								if err != nil {
									log.Printf("创建文件失败: %v", err)
									return
								}
								defer file.Close()

								// 写入表头
								file.WriteString("助记词---私钥---地址\n")

								for i := 0; i < num; i++ {
									wallet, err := generateWallet()
									if err != nil {
										log.Printf("生成钱包失败: %v", err)
										continue
									}

									walletData := Wallet{
										Index:    i + 1,
										Mnemonic: wallet.Mnemonic,
										Private:  wallet.PrivateKey,
										Address:  wallet.PublicKey,
									}

									// 简化的文件输出格式
									line := fmt.Sprintf("%s---%s---%s\n",
										wallet.Mnemonic, wallet.PrivateKey, wallet.PublicKey)
									file.WriteString(line)

									mw.Synchronize(func() {
										model.items = append(model.items, walletData)
										model.PublishRowsReset()
										progressBar.SetValue(int(float64(i+1) / float64(num) * 100))
									})
								}

								mw.Synchronize(func() {
									walk.MsgBox(mw, "提示",
										fmt.Sprintf("已成功生成 %d 个钱包!\n保存至文件: %s", len(allWallets), fileName),
										walk.MsgBoxIconInformation)
								})
							}()
						},
					},
				},
			},
			ProgressBar{
				AssignTo: &progressBar,
				MinValue: 0,
				MaxValue: 100,
			},
			TableView{
				AssignTo: &tableView,
				Model:    model,
				Columns: []TableViewColumn{
					{Title: "序号", Width: 60},
					{Title: "助记词", Width: 300},
					{Title: "私钥", Width: 300},
					{Title: "地址", Width: 500},
				},
				StyleCell: func(style *walk.CellStyle) {
					if style.Row()%2 == 0 {
						style.BackgroundColor = walk.RGB(245, 245, 245)
					}
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
		return
	}
}
