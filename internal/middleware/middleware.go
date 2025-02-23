package middleware

import (
	"compress/gzip"
	"fmt"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/models/response"
	"go-musthave-diploma-tpl/internal/session"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	logKeyError = "error"
)

type gzipProvider struct {
	writer gin.ResponseWriter
	reader io.ReadCloser
}
type gzipWriter struct {
	gin.ResponseWriter
	zw *gzip.Writer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	n, err := w.zw.Write(b)
	if err != nil {
		return 0, fmt.Errorf("failed to write compressed data: %w", err)
	}
	return n, nil
}
func (w *gzipWriter) Close() error {
	err := w.zw.Close()
	if err != nil {
		return fmt.Errorf("failed to close compressed data: %w", err)
	}
	return nil
}
func (p *gzipProvider) gzipHandler() *gzip.Writer {
	zw := gzip.NewWriter(p.writer)
	return zw
}
func (p *gzipProvider) unGzipHandler(sgr *zap.SugaredLogger) (*gzip.Reader, error) {
	zr, err := gzip.NewReader(p.reader)
	if err != nil {
		sgr.Errorw(
			"gzip middleware reader failed",
			logKeyError, err.Error(),
		)

		return nil, fmt.Errorf("failed to ungzip request body: %w", err)
	}

	return zr, nil
}
func GzipMiddleware(sgr *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.Request.Header.Get("Content-Type")
		supportsJSON := strings.Contains(contentType, "application/json")
		supportsHTML := strings.Contains(contentType, "text/html")

		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip && (supportsJSON || supportsHTML) {
			p := gzipProvider{
				writer: c.Writer,
			}
			zw := p.gzipHandler()
			defer func() {
				err := zw.Close()
				if err != nil {
					sgr.Errorw(
						"gzip middleware write close failed",
						logKeyError, err.Error(),
					)
				}
			}()

			c.Writer = &gzipWriter{
				ResponseWriter: p.writer,
				zw:             zw,
			}
			c.Writer.Header().Set("Content-Encoding", "gzip")
		}

		contentEncoding := c.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			p := gzipProvider{
				reader: c.Request.Body,
			}
			zr, err := p.unGzipHandler(sgr)
			defer func() {
				err := zr.Close()
				if err != nil {
					sgr.Errorw(
						"gzip middleware reader close failed",
						logKeyError, err.Error(),
					)
				}
			}()

			if err != nil {
				sgr.Errorw(
					"gzip middleware reader failed",
					logKeyError, err.Error(),
				)
				c.Writer.WriteHeader(http.StatusBadRequest)
				_, err = c.Writer.WriteString(http.StatusText(http.StatusBadRequest))
				if err != nil {
					sgr.Errorw(
						"gzip middleware write failed",
						logKeyError, err.Error(),
					)
				}
				c.Abort()
				return
			}

			c.Request.Body = io.NopCloser(zr)
		}

		c.Next()
	}
}

func AuthMiddleware(sgr *zap.SugaredLogger, cnf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ssp := session.SessionProvider{
			Config: cnf,
		}
		err := ssp.CheckToken(c)
		if err != nil {
			sgr.Errorw(
				"failed to verify token",
				"info", "incorrect access token",
			)

			message := response.Response{
				Message: "Невалидный токен",
			}

			c.JSON(http.StatusUnauthorized, message)
			c.Abort()
			return
		}

		c.Next()
	}
}
