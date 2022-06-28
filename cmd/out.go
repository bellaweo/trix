package cmd

import (
	matrix "github.com/matrix-org/gomatrix"
	"github.com/spf13/cobra"
)

var (
	text string // text to send to the matrix room

	out = &cobra.Command{
		Use:   "out",
		Short: "send text to the matrix channel",
		Long:  `send string output to the matrix channel`,
		Run: func(cmd *cobra.Command, args []string) {
			vars()

			cli, _ := matrix.NewClient(host, "", "")
			resp, err := cli.Login(&matrix.ReqLogin{
				Type:     "m.login.password",
				User:     user,
				Password: pass,
			})
			if err != nil {
				panic(err)
			}
			cli.SetCredentials(resp.UserID, resp.AccessToken)
			_, _ = cli.SendNotice(room, text)

		},
	}
)

func init() {

	out.Flags().StringVarP(&text, "text", "t", "", "text to send to the matrix room")
	out.MarkFlagRequired("text")

	rootCmd.AddCommand(out)
}
