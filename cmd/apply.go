/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"gthub.com/Mrliuch/cd-tools/pkg/apply"
	"gthub.com/Mrliuch/cd-tools/pkg/utils"

	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "创建或升级集群内的Deployment",
	Long: `创建或升级集群内的Deployment，如果集群内已经存在则进行升级，同时保留原始yaml文件，如不存在，则进行创建。
失败后会回滚到之前版本。
注意目前仅支持Deployment类型。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workdir, _ := cmd.Flags().GetString("workdir")
		err := utils.PathExistsOrCreate(workdir)
		if err != nil {
			logrus.Fatalf("创建工作目录失败：%s", err.Error())
		}
		return apply.Apply(cmd, workdir)
	},
}

func init() {
	applyCmd.Flags().StringP("kube-config", "c", "/root/.kube/config", "指定KubeConfig文件")
	applyCmd.Flags().StringP("workdir", "w", "/root/.cd-tools", "工作路径，用来存放备份文件")
	applyCmd.Flags().StringP("file", "f", "", "指定创建或更新的K8S yaml文件路径")
	applyCmd.Flags().IntP("timeout", "t", 60, "指定超时时间，单位s，最小设置为5")
	applyCmd.Flags().StringP("cluster-name", "n", "", "设置集群名称，默认为master ip")
	rootCmd.AddCommand(applyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// applyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// applyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
