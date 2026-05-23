package storage

import "testing"

func TestValidateImageAllowsPNGBySniffedMIME(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R'}

	mimeType, extension, err := validateImage(data, int64(len(data)))
	if err != nil {
		t.Fatalf("validateImage error = %v", err)
	}
	if mimeType != MIMEPNG || extension != ".png" {
		t.Fatalf("mime/ext = %s/%s, want %s/.png", mimeType, extension, MIMEPNG)
	}
}

func TestValidateImageRejectsInvalidMIME(t *testing.T) {
	if _, _, err := validateImage([]byte("not an image"), 1024); err == nil {
		t.Fatal("validateImage error = nil, want invalid MIME rejection")
	}
}

func TestValidateImageRejectsOversizedFile(t *testing.T) {
	data := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

	if _, _, err := validateImage(data, int64(len(data)-1)); err == nil {
		t.Fatal("validateImage error = nil, want oversized rejection")
	}
}

func TestValidateFolderSegmentRejectsPathTraversal(t *testing.T) {
	for _, folder := range []string{"../products", `products\evil`, ".", "..", ""} {
		if err := validateFolderSegment(folder); err == nil {
			t.Fatalf("validateFolderSegment(%q) error = nil, want rejection", folder)
		}
	}
}
