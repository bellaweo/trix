package main

import "codeberg.org/meh/trix/cmd"

func main() {
	//	cli, _ := gomatrix.NewClient("https://matrix.org", "@betch:matrix.org", "syt_YmV0Y2g_yZhXzdlyiUzOMSzZcAMt_3Lzy1K")
	//	_, _ = cli.SendText("!MzZofkzZHwgsixQqgk:matrix.org", "Down the rabbit hole")

	//	host := os.Getenv("TRIX_HOST")
	//	user := os.Getenv("TRIX_USER")
	//	pass := os.Getenv("TRIX_PASS")
	//	room := os.Getenv("TRIX_ROOM")

	//	cli, _ := matrix.NewClient(host, "", "")
	//	resp, err := cli.Login(&matrix.ReqLogin{
	//		Type:     "m.login.password",
	//		User:     user,
	//		Password: pass,
	//	})
	//	if err != nil {
	//		panic(err)
	//	}
	//	cli.SetCredentials(resp.UserID, resp.AccessToken)
	//
	//	_, _ = cli.SendNotice(room, "Down the rabbit hole")

	cmd.Execute()

}
