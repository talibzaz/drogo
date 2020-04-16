package cmd

import (
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"github.com/kataras/iris"
	"github.com/spf13/viper"
	"code.drogo.gw.com/http"
)

var isProd = false

var startCmd = &cobra.Command{
	Use:  "roar",
	Long: "starts the drogo server",
	PreRun: func(cmd *cobra.Command, args []string) {

		log.SetFormatter(&log.JSONFormatter{})
		viper.Set("DROGO_SERVER", ":80")
		viper.AutomaticEnv()
	},
	Run: func(cmd *cobra.Command, args []string) {
		http.NewDrogoServer(isProd).Run(iris.Addr(viper.GetString("DROGO_SERVER")))
	},
}
