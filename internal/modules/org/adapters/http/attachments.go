package http

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/internal/platform/uploads"
)

func registerAttachmentRoutes(r chi.Router, attachments ports.AttachmentService, authorizer authx.Authorizer, uploadsDir string) {
	r.Get("/request-attachments", listRequestAttachments(attachments, authorizer))
	r.Post("/request-attachments", uploadRequestAttachment(attachments, authorizer, uploadsDir))
	r.Get("/request-attachments/{id}/download", downloadRequestAttachment(attachments, authorizer))
}

func listRequestAttachments(attachments ports.AttachmentService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resourceType := r.URL.Query().Get("resourceType")
		resourceIDRaw := r.URL.Query().Get("resourceId")
		if resourceType == "" || resourceIDRaw == "" {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "resourceType and resourceId required")
			return
		}
		module := domain.ResourceModule(resourceType)
		if module == "" || !authorizer.Can(r.Context(), authx.Module(module), authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		resourceID, err := uuid.Parse(resourceIDRaw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid resourceId")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := attachments.List(r.Context(), identity.TenantID, resourceType, resourceID)
		if err != nil {
			writeAttachmentError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func uploadRequestAttachment(attachments ports.AttachmentService, authorizer authx.Authorizer, uploadsDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(uploads.MaxAttachmentBytes + (1 << 20)); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid multipart form")
			return
		}
		resourceType := r.FormValue("resourceType")
		resourceIDRaw := r.FormValue("resourceId")
		module := domain.ResourceModule(resourceType)
		if module == "" || !authorizer.Can(r.Context(), authx.Module(module), authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		resourceID, err := uuid.Parse(resourceIDRaw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid resourceId")
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "file required")
			return
		}
		defer file.Close()
		identity, _ := authx.FromContext(r.Context())
		att, err := attachments.Create(r.Context(), ports.CreateAttachmentCommand{
			TenantID:     identity.TenantID,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			FileName:     header.Filename,
			MimeType:     header.Header.Get("Content-Type"),
			Content:      file,
			UploadedBy:   identity.UserID,
			UploadsDir:   uploadsDir,
		})
		if err != nil {
			writeAttachmentUploadError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusCreated, att)
	}
}

func downloadRequestAttachment(attachments ports.AttachmentService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		att, err := attachments.Get(r.Context(), identity.TenantID, id)
		if err != nil {
			if errors.Is(err, domain.ErrAttachmentNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		module := domain.ResourceModule(att.ResourceType)
		if module == "" || !authorizer.Can(r.Context(), authx.Module(module), authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		path, ok := uploads.AttachmentPath(att.StoragePath)
		if !ok {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "file not found")
			return
		}
		w.Header().Set("Content-Type", att.MimeType)
		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(att.FileName))
		http.ServeFile(w, r, path)
	}
}

func writeAttachmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAttachmentResourceNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidAttachmentTarget):
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}

func writeAttachmentUploadError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, uploads.ErrInvalidAttachment),
		errors.Is(err, uploads.ErrAttachmentTooLarge),
		errors.Is(err, uploads.ErrUnsupportedExt),
		errors.Is(err, domain.ErrInvalidAttachmentTarget),
		errors.Is(err, domain.ErrAttachmentResourceNotFound):
		if errors.Is(err, domain.ErrAttachmentResourceNotFound) {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}
