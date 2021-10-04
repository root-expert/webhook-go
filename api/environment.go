package api

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/voxpupuli/webhook-go/config"
	"github.com/voxpupuli/webhook-go/lib/helpers"
	"github.com/voxpupuli/webhook-go/lib/parsers"
)

type EnvironmentController struct{}

func (e EnvironmentController) DeployEnvironment(c *gin.Context) {
	data := parsers.Data{}
	h := helpers.Helper{}
	cmd := exec.Command("r10k", "deploy", "environment")
	conf := config.GetConfig().R10k
	prefix := h.GetPrefix(data, conf.Prefix)
	branch := h.GetBranch(data, conf.DefaultBranch)
	env := h.GetEnvironment(branch, prefix, conf.AllowUppercase)

	err := data.ParseData(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error Parsing Webhook", "error": err})
		log.Errorf("error parsing webhook: %s", err)
		c.Abort()
		return
	}

	if !data.Completed {
		c.JSON(http.StatusAccepted, gin.H{"message": "Received webhook but job is not complete. Ignoring."})
		log.Debug("Received webhook but job is not complete. Ignoring...")
		c.Abort()
		return
	} else if !data.Succeed {
		c.JSON(http.StatusAccepted, gin.H{"message": "Received webhook but job failed. Ignoring."})
		log.Debug("Received webhook but job failed. Ignoring...")
		c.Abort()
		return
	}

	cmd.Args = append(cmd.Args, env)

	cmd.Args = append(cmd.Args, fmt.Sprintf("-c %s", h.GetR10kConfig()))

	if conf.Verbose {
		cmd.Args = append(cmd.Args, "-v")
	}

	res, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("cmd.Run() failed with error %s", string(res))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error executing command", "error": string(res)})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": string(res)})
	log.Info(fmt.Sprintf("\n%s", string(res)))
}
