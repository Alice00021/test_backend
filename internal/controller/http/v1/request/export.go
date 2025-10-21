package request

import "mime/multipart"

type UploadFileRequest struct {
	File *multipart.FileHeader `form:"file" validate:"required"`
}
