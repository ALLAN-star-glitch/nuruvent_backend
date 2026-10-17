// internal/modules/certificate/service/template_media.go

package service

import "context"

func (s *serviceImpl) UploadTemplateImage(
	ctx context.Context,
	templateID, filename, contentType string,
	data []byte,
) (string, error) {
	return s.deps.Storage.UploadTemplateImage(ctx, templateID, filename, contentType, data)
}

func (s *serviceImpl) DeleteTemplateImage(ctx context.Context, url string) error {
	return s.deps.Storage.DeleteFileByURL(ctx, url)
}