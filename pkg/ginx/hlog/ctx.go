package hlog

import (
	"github.com/gin-gonic/gin"
)

// Attrs carries map. It implements value for that key and
// delegates all other calls to the embedded Context.
type Attrs map[string]interface{}

// PutAttr put an attribute into the Attributes in the context.
func PutAttr(c *gin.Context, key string, value interface{}) {
	c.Set(key, value)
}

// PutAttrMap put an attribute map into the Attributes in the context.
func PutAttrMap(c *gin.Context, attrs map[string]interface{}) {
	for k, v := range attrs {
		c.Set(k, v)
	}
}
