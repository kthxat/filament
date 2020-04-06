package frontend

import (
	"archive/zip"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/kthxat/filament/backends"

	"github.com/BurntSushi/toml"
	rice "github.com/GeertJohan/go.rice"
	humanize "github.com/dustin/go-humanize"
	gintemplate "github.com/foolin/gin-template"
	"github.com/foolin/gin-template/supports/gorice"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"golang.org/x/text/language"

	"github.com/kthxat/filament/app"
	"github.com/kthxat/filament/config"
)

const (
	relPathActions         = ".filament"
	relPathArchiveZip      = relPathActions + "/archive.zip"
	relPathArchiveTar      = relPathActions + "/archive.tar"
	relPathArchiveTarXZ    = relPathArchiveTar + ".xz"
	relPathArchiveTarGZip  = relPathArchiveTar + ".gz"
	relPathArchiveTarBZip2 = relPathArchiveTar + ".bz2"
	relPathArchiveTar7Zip  = relPathArchiveTar + ".7z"
)

type FrontendServer struct {
	httpServer *http.Server
	i18n       *i18n.Bundle
}

type fileMapping struct {
	Path     string
	FileInfo os.FileInfo
}

func NewFrontendServer(config *config.HTTPConfig) *FrontendServer {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

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

		lang := session.Language()
		accept := c.GetHeader("Accept-Language")
		localizer := i18n.NewLocalizer(bundle, lang, accept)

		relpath := c.Param("path")

		switch {
		case strings.HasSuffix(relpath, "/"+relPathArchiveZip):
			relpath = strings.TrimSuffix(relpath, relPathArchiveZip)

			fileInfo, err := session.Storage().Stat(relpath)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			if !fileInfo.IsDir() {
				c.AbortWithStatus(http.StatusConflict)
				return
			}

			mappings := []fileMapping{}
			relpathFilepath := filepath.FromSlash(relpath)
			err = backends.ReadDirRecursively(session.Storage(), relpath, func(pwd string, fi os.FileInfo, err error) error {
				log.Printf("I % -99s %s", path.Join(pwd, fi.Name()), fi.ModTime())
				if err != nil {
					return err
				}
				pwdFilepath := filepath.FromSlash(pwd)
				recalculatedPath, err := filepath.Rel(relpathFilepath, pwdFilepath)
				if err != nil {
					return err
				}
				mappings = append(mappings, fileMapping{
					Path: path.Join(
						filepath.ToSlash(recalculatedPath),
						fi.Name()),
					FileInfo: fi,
				})
				return nil
			})
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			sort.Slice(mappings, func(i, j int) bool {
				a := mappings[i]
				b := mappings[j]
				// 1. directories first
				// 2. alphabetical sorting (a to z)
				return a.FileInfo.IsDir() &&
					strings.Compare(a.Path, b.Path) == -1
			})

			c.Header("content-type", "application/zip")
			c.Writer.WriteHeaderNow()

			z := zip.NewWriter(c.Writer)
			// dotFH := &zip.FileHeader{
			// 	Name:               "",
			// 	UncompressedSize64: uint64(fileInfo.Size()),
			// }
			// dotFH.SetModTime(fileInfo.ModTime())
			// dotFH.SetMode(fileInfo.Mode())
			// z.CreateHeader(dotFH)
			for _, file := range mappings {
				fh, err := zip.FileInfoHeader(file.FileInfo)
				if err != nil {
					c.Error(err)
					return
				}
				fh.Name = file.Path
				fh.SetModTime(file.FileInfo.ModTime())

				if !file.FileInfo.IsDir() {
					fh.Method = zip.Deflate
				}

				log.Printf("O % -99s %s", fh.Name, fh.ModTime())

				zw, err := z.CreateHeader(fh)
				if file.FileInfo.IsDir() {
					continue
				}
				if err != nil {
					z.SetComment("Incomplete file")
					c.Error(err)
					return
				}

				err = session.Storage().Retrieve(path.Join(relpath, file.Path), zw)
				if err != nil {
					z.SetComment("Incomplete file")
					c.Error(err)
					return
				}
			}

			z.Close()
			return

		case
			strings.HasSuffix(relpath, "/"+relPathArchiveTar),
			strings.HasSuffix(relpath, "/"+relPathArchiveTarBZip2),
			strings.HasSuffix(relpath, "/"+relPathArchiveTarGZip),
			strings.HasSuffix(relpath, "/"+relPathArchiveTarXZ),
			strings.HasSuffix(relpath, "/"+relPathArchiveTar7Zip):
			c.AbortWithStatus(http.StatusNotImplemented)
			return
		case strings.HasSuffix(relpath, "/.filament/archive.7z"):
			c.AbortWithStatus(http.StatusNotImplemented)
			return
		}

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

			localizedDownloadAsArchiveZIP, err := localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "DownloadAsArchiveZIP",
					Other: "Download as ZIP archive",
				},
			})
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
			}

			data := gin.H{
				"Path":  relpath,
				"Files": files,
				"Actions": []gin.H{
					gin.H{
						"Name": localizedDownloadAsArchiveZIP,
						"Link": relPathArchiveZip,
					},
				},
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
		i18n:       bundle,
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
