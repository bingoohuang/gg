package anyfn

import (
	"net/http"

	"github.com/bingoohuang/gg/pkg/ginx"
	"github.com/gin-gonic/gin"
)

// DirectResponse represents the direct response.
type DirectResponse struct {
	Code        int
	Error       error
	JSON        interface{}
	String      string
	ContentType string
	Header      map[string]string
}

func (d DirectResponse) Deal(c *gin.Context) {
	if d.Code == 0 {
		if d.Error != nil {
			d.Code = http.StatusInternalServerError
		} else {
			d.Code = http.StatusOK
		}
	}

	if d.ContentType != "" {
		c.Header("Content-Type", d.ContentType)
	}

	for k, v := range d.Header {
		c.Header(k, v)
	}

	if d.Error != nil {
		errString := d.Error.Error()
		c.String(d.Code, errString)
		return
	}

	if d.JSON != nil {
		c.Render(d.Code, ginx.JSONRender{Data: d.JSON})
		return
	}

	if d.String != "" {
		c.String(d.Code, d.String)
		return
	}

	c.Status(d.Code)
}
