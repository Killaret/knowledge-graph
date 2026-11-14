package notehandler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	importer "knowledge-graph/internal/application/import"
	apicommon "knowledge-graph/internal/interfaces/api/common"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
)

// maxDocumentFileSize caps the uploaded document. The contract (TZ §3) sets no
// limit; 8 MiB follows from client_max_body_size 10m in nginx.conf and
// nginx.personal.conf, leaving headroom for the multipart envelope.
const maxDocumentFileSize = 8 << 20

// ImportDocumentAvailability reports whether the Java source-text-handler
// listener is alive: GET /api/v1/import/document/availability.
func (h *Handler) ImportDocumentAvailability(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "import_document_availability")

	if _, exists := middleware.GetUserID(c); !exists {
		apicommon.Unauthorized(c)
		return
	}

	available := false
	if h.importSvc != nil {
		if ok, err := h.importSvc.DocumentAvailability(c.Request.Context()); err == nil {
			available = ok
		} else {
			log.Printf("[NoteHandler] document availability check failed: %v", err)
		}
	}
	apicommon.JSON(c, http.StatusOK, gin.H{"available": available})
}

// ImportDocument accepts a document for the Java source-text-handler:
// POST /api/v1/import/document (multipart/form-data, fields type, content,
// file). With no live listener it answers 503 and queues nothing.
func (h *Handler) ImportDocument(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "import_document_create")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		apicommon.Unauthorized(c)
		return
	}

	if h.importSvc == nil {
		apicommon.InternalErrorWithMessage(c, "Import service is not configured")
		return
	}

	alive, err := h.importSvc.DocumentAvailability(c.Request.Context())
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Failed to check document handler availability")
		return
	}
	if !alive {
		apicommon.Error(c, http.StatusServiceUnavailable, "source_text_handler_unavailable",
			"Document handler is unavailable")
		return
	}

	jwt := extractRequestToken(c)
	if jwt == "" {
		apicommon.Unauthorized(c)
		return
	}

	input, ferr := parseDocumentForm(c)
	if ferr != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("request", apicommon.ReasonInvalidValue, ferr.Error()),
		})
		return
	}

	taskID, err := h.importSvc.SubmitDocument(c.Request.Context(), userID, jwt, input)
	if err != nil {
		if errors.Is(err, importer.ErrDocumentUnavailable) {
			apicommon.Error(c, http.StatusServiceUnavailable, "source_text_handler_unavailable",
				"Document handler is unavailable")
			return
		}
		apicommon.BadRequestSimple(c, err.Error())
		return
	}

	apicommon.JSON(c, http.StatusAccepted, gin.H{"task_id": taskID})
}

// extractRequestToken returns the bearer token Java will present to
// /import/batch — the same credential the user authenticated with.
func extractRequestToken(c *gin.Context) string {
	if header := c.GetHeader("Authorization"); header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}
	if cookie, err := c.Cookie("access_token"); err == nil {
		return cookie
	}
	return ""
}

// parseDocumentForm reads the multipart form into a validated DocumentInput.
// TEXT needs a non-empty content, URL additionally passes the importer SSRF
// policy, FILE is bounded by maxDocumentFileSize.
func parseDocumentForm(c *gin.Context) (importer.DocumentInput, error) {
	docType := strings.ToUpper(strings.TrimSpace(c.PostForm("type")))

	switch docType {
	case "TEXT":
		content := c.PostForm("content")
		if strings.TrimSpace(content) == "" {
			return importer.DocumentInput{}, errors.New("content must not be empty")
		}
		return importer.DocumentInput{Type: "TEXT", Content: content}, nil

	case "URL":
		rawURL := strings.TrimSpace(c.PostForm("content"))
		if rawURL == "" {
			return importer.DocumentInput{}, errors.New("content must be a URL")
		}
		if !importer.IsAllowedURL(rawURL) {
			return importer.DocumentInput{}, errors.New("URL is not allowed or uses an unsupported scheme")
		}
		return importer.DocumentInput{Type: "URL", Content: rawURL}, nil

	case "FILE":
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return importer.DocumentInput{}, errors.New("file is required")
		}
		if fileHeader.Size > maxDocumentFileSize {
			return importer.DocumentInput{}, fmt.Errorf("file is larger than %d MiB", maxDocumentFileSize>>20)
		}
		f, err := fileHeader.Open()
		if err != nil {
			return importer.DocumentInput{}, errors.New("failed to read uploaded file")
		}
		defer f.Close()
		content, err := io.ReadAll(io.LimitReader(f, maxDocumentFileSize+1))
		if err != nil {
			return importer.DocumentInput{}, errors.New("failed to read uploaded file")
		}
		if len(content) > maxDocumentFileSize {
			return importer.DocumentInput{}, fmt.Errorf("file is larger than %d MiB", maxDocumentFileSize>>20)
		}
		return importer.DocumentInput{
			Type:        "FILE",
			FileContent: content,
			ContentType: fileHeader.Header.Get("Content-Type"),
			Filename:    fileHeader.Filename,
		}, nil

	default:
		return importer.DocumentInput{}, errors.New("type must be one of TEXT, URL, FILE")
	}
}
