package anyfn

import (
	"mime"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// DlFile represents the file to be downloaded.
type DlFile struct {
	DiskFile string
	Filename string
	Content  []byte
}

// Deal is the processor for a specified type.
func (d DlFile) Deal(c *gin.Context) {
	c.Header("Content-Disposition", d.createContentDisposition())
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	if d.DiskFile != "" {
		c.File(d.DiskFile)
		return
	}

	_, _ = c.Writer.Write(d.Content)
}

func (d DlFile) createContentDisposition() string {
	m := map[string]string{"filename": d.getDownloadFilename()}
	return mime.FormatMediaType("attachment", m)
}

func (d DlFile) getDownloadFilename() string {
	filename := d.Filename

	if filename == "" {
		filename = filepath.Base(d.DiskFile)
	}

	if filename == "" {
		return "dl"
	}

	return filename
}
