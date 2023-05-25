package main

func main() {
	bot, err := NewMyBot()
	if err != nil {
		panic(err)
	}
	bot.Run()
}
