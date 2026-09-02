package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type APIKeyGroupResolver interface {
	Resolve(c *gin.Context, apiKey *service.APIKey) (*service.APIKey, error)
}

type smartRouteGroupResolver struct {
	service    *service.SmartRouteService
	imageTasks *service.ImageTaskService
}

var smartRouteClaudeCodeValidator = service.NewClaudeCodeValidator()

func NewSmartRouteGroupResolver(svc *service.SmartRouteService, imageTasks *service.ImageTaskService) APIKeyGroupResolver {
	return &smartRouteGroupResolver{service: svc, imageTasks: imageTasks}
}

func abortSmartRouteResolveError(c *gin.Context, err error) {
	status := infraerrors.Code(err)
	reason := infraerrors.Reason(err)
	message := infraerrors.Message(err)
	if reason != "MODEL_NOT_FOUND" {
		AbortWithError(c, status, reason, message)
		return
	}
	if strings.Contains(strings.ToLower(c.Request.URL.Path), "/messages") {
		c.JSON(status, gin.H{"type": "error", "error": gin.H{"type": "not_found_error", "message": message}})
	} else {
		c.JSON(status, gin.H{"error": gin.H{"type": "invalid_request_error", "code": "model_not_found", "param": "model", "message": message}})
	}
	c.Abort()
}

func (r *smartRouteGroupResolver) Resolve(c *gin.Context, apiKey *service.APIKey) (*service.APIKey, error) {
	if r == nil || r.service == nil || c == nil || c.Request == nil {
		return apiKey, nil
	}
	descriptor, err := smartRouteRequestDescriptor(c.Request)
	if err != nil {
		return nil, err
	}
	if descriptor.ClaudeCode {
		c.Request = c.Request.WithContext(service.SetClaudeCodeClient(c.Request.Context(), true))
	}
	if isAsyncImageTaskRead(c.Request.Method, c.Request.URL.Path) && r.imageTasks != nil {
		taskID := strings.TrimPrefix(c.Request.URL.Path, "/v1/images/tasks/")
		taskID = strings.TrimPrefix(taskID, "/images/tasks/")
		groupID, taskErr := r.imageTasks.RoutingGroup(c.Request.Context(), service.ImageTaskOwner{
			UserID: apiKey.UserID, APIKeyID: apiKey.ID,
		}, taskID)
		if taskErr == nil && groupID > 0 {
			resolved, config, resolveErr := r.service.ResolveRecordedGroup(c.Request.Context(), apiKey, groupID)
			if config != nil {
				c.Set(string(ContextKeySmartRoute), config)
			}
			return resolved, resolveErr
		}
	}
	resolved, config, err := r.service.Resolve(c.Request.Context(), apiKey, descriptor)
	if config != nil {
		c.Set(string(ContextKeySmartRoute), config)
	}
	return resolved, err
}

func smartRouteRequestDescriptor(req *http.Request) (service.SmartRouteRequest, error) {
	descriptor := service.SmartRouteRequest{Method: req.Method, Path: req.URL.Path}
	path := strings.ToLower(req.URL.Path)
	if strings.Contains(path, "/images/batches") {
		descriptor.Kind = "batch_image"
	} else if strings.Contains(path, "/images") {
		descriptor.Kind = "image"
	} else {
		descriptor.Kind = "text"
	}
	if idx := strings.LastIndex(path, "/models/"); idx >= 0 {
		modelPart := req.URL.Path[idx+len("/models/"):]
		if colon := strings.IndexByte(modelPart, ':'); colon >= 0 {
			modelPart = modelPart[:colon]
		}
		descriptor.Model = modelPart
	}
	contentType := req.Header.Get("Content-Type")
	if req.Body == nil {
		descriptor.ClaudeCode = smartRouteClaudeCodeValidator.Validate(req, nil)
		return descriptor, nil
	}
	mediaType, params, _ := mime.ParseMediaType(contentType)
	mediaType = strings.ToLower(mediaType)
	if !strings.Contains(mediaType, "json") && mediaType != "multipart/form-data" {
		return descriptor, nil
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return descriptor, err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	if mediaType == "multipart/form-data" {
		descriptor.Model = smartRouteMultipartModel(body, params["boundary"])
		descriptor.ClaudeCode = smartRouteClaudeCodeValidator.Validate(req, nil)
		return descriptor, nil
	}
	if descriptor.Model == "" {
		descriptor.Model = strings.TrimSpace(gjson.GetBytes(body, "model").String())
	}
	descriptor.Streaming = gjson.GetBytes(body, "stream").Bool()
	var bodyMap map[string]any
	if json.Unmarshal(body, &bodyMap) == nil {
		descriptor.ClaudeCode = smartRouteClaudeCodeValidator.Validate(req, bodyMap)
	}
	return descriptor, nil
}

func smartRouteMultipartModel(body []byte, boundary string) string {
	if boundary == "" {
		return ""
	}
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := reader.NextPart()
		if err != nil {
			return ""
		}
		if part.FormName() != "model" {
			_ = part.Close()
			continue
		}
		value, err := io.ReadAll(io.LimitReader(part, 4096))
		_ = part.Close()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(value))
	}
}
