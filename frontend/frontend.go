package frontend

import (
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	humanize "github.com/dustin/go-humanize"
	gintemplate "github.com/foolin/gin-template"
	"github.com/foolin/gin-template/supports/gorice"
	"github.com/gin-gonic/gin"
	"github.com/kthxat/filament/app"
	"github.com/kthxat/filament/config"

)

type FrontendServer struct {
	httpServer *http.Server
}

func NewFrontendServer(config *config.HTTPConfig) *FrontendServer {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Session management
	authorized := r.Group("/", UsernameBasedSessions(config.AuthenticationRealm))

	// Templates via rice box
	r.HTMLRender = gorice.NewWithConfig(rice.MustFindBox("templates"), gintemplate.TemplateConfig{
		Funcs: template.FuncMap{
			"humanize_bytes": func(bytes int64) string {
				return humanize.Bytes(uint64(bytes))
			},
		},
	})

	// Routes
	authorized.GET("/*path", func(c *gin.Context) {
		session := app.GetSessionById(c.GetString(gin.AuthUserKey))
		if session == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		session.Increment()
		defer session.Decrement()

		relpath := c.Param("path")
		fileInfo, err := session.Storage().Stat(relpath)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if fileInfo.IsDir() {
			if !strings.HasSuffix(relpath, "/") {
				c.Redirect(http.StatusTemporaryRedirect, relpath+"/")
				return
			}
			files, err := session.Storage().ReadDir(relpath)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			data := gin.H{
				"Path":  relpath,
				"Files": files,
			}
			if path.Base(relpath) != path.Clean(relpath) {
				data["ParentPath"] = ".."
			}
			sort.Slice(files, func(i, j int) bool {
				a := files[i]
				b := files[j]
				// 1. directories first
				// 2. alphabetical sorting (a to z)
				return a.IsDir() &&
					strings.Compare(a.Name(), b.Name()) == -1
			})
			// c.JSON(http.StatusOK, files)
			c.HTML(http.StatusOK, "directory.html", data)
			return
		}

		c.Header("content-length", fmt.Sprintf("%d", fileInfo.Size()))

		if mimeType := mime.TypeByExtension(path.Ext(relpath)); len(mimeType) > 0 {
			c.Header("content-type", mimeType)
		} else {
			c.Header("content-type", "application/octet-stream")
		}

		c.Writer.WriteHeaderNow()

		err = session.Storage().Retrieve(relpath, c.Writer)
		if err != nil {
			log.Printf("Writing file from storage to HTTP failed: %s",
				err.Error())
		}

		return
	})

	httpServer := new(http.Server)
	httpServer.Addr = config.ListenAddress
	httpServer.Handler = r

	return &FrontendServer{
		httpServer: httpServer,
	}
}

func (f *FrontendServer) ListenAndServe() error {
	return f.httpServer.ListenAndServe()
}

func (f *FrontendServer) Close() error {
	return f.httpServer.Close()
}

func UsernameBasedSessions(realm string) gin.HandlerFunc {
	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)
	return func(c *gin.Context) {
		var sid string
		if username, password, ok := parseBasicAuth(c.Request.Header.Get("Authorization")); ok {
			// Valid Authentication header was passed!
			sid = app.Authenticate(username, password)
		}

		if len(sid) <= 0 {
			// Credentials doesn't match, we return 401 and abort handlers chain.
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(gin.AuthUserKey, sid)
	}
}
