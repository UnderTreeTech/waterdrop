/*
 *
 * Copyright 2021 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package store

import (
	"mime"
	"strings"
	"time"
)

// mimeTypes mime types
var mimeTypes = map[string]string{
	".avif":   "image/avif",
	".css":    "text/css; charset=utf-8",
	".gif":    "image/gif",
	".htm":    "text/html; charset=utf-8",
	".html":   "text/html; charset=utf-8",
	".jpeg":   "image/jpeg",
	".jpg":    "image/jpeg",
	".js":     "text/javascript; charset=utf-8",
	".json":   "application/json",
	".mjs":    "text/javascript; charset=utf-8",
	".pdf":    "application/pdf",
	".png":    "image/png",
	".svg":    "image/svg+xml",
	".wasm":   "application/wasm",
	".webp":   "image/webp",
	".xml":    "text/xml; charset=utf-8",
	".cr2":    "image/x-canon-cr2",
	".tif":    "image/tiff",
	".bmp":    "image/bmp",
	".heif":   "image/heif",
	".jxr":    "image/vnd.ms-photo",
	".psd":    "image/vnd.adobe.photoshop",
	".ico":    "image/vnd.microsoft.icon",
	".dwg":    "image/vnd.dwg",
	".mp4":    "video/mp4",
	".m4v":    "video/x-m4v",
	".mkv":    "video/x-matroska",
	".webm":   "video/webm",
	".mov":    "video/quicktime",
	".avi":    "video/x-msvideo",
	".wmv":    "video/x-ms-wmv",
	".mpg":    "video/mpeg",
	".flv":    "video/x-flv",
	".3gp":    "video/3gpp",
	".mid":    "audio/midi",
	".mp3":    "audio/mpeg",
	".m4a":    "audio/m4a",
	".ogg":    "audio/ogg",
	".flac":   "audio/x-flac",
	".wav":    "audio/x-wav",
	".amr":    "audio/amr",
	".aac":    "audio/aac",
	".epub":   "application/epub+zip",
	".zip":    "application/zip",
	".tar":    "application/x-tar",
	".rar":    "application/vnd.rar",
	".gz":     "application/gzip",
	".bz2":    "application/x-bzip2",
	".7z":     "application/x-7z-compressed",
	".xz":     "application/x-xz",
	".zstd":   "application/zstd",
	".exe":    "application/vnd.microsoft.portable-executable",
	".swf":    "application/x-shockwave-flash",
	".rtf":    "application/rtf",
	".iso":    "application/x-iso9660-image",
	".eot":    "application/octet-stream",
	".ps":     "application/postscript",
	".sqlite": "application/vnd.sqlite3",
	".nes":    "application/x-nintendo-nes-rom",
	".crx":    "application/x-google-chrome-extension",
	".cab":    "application/vnd.ms-cab-compressed",
	".deb":    "application/vnd.debian.binary-package",
	".ar":     "application/x-unix-archive",
	".Z":      "application/x-compress",
	".lz":     "application/x-lzip",
	".rpm":    "application/x-rpm",
	".elf":    "application/x-executable",
	".dcm":    "application/dicom",
	".doc":    "application/msword",
	".docx":   "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":    "application/vnd.ms-excel",
	".xlsx":   "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":    "application/vnd.ms-powerpoint",
	".pptx":   "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".woff":   "application/font-woff",
	".woff2":  "application/font-woff",
	".ttf":    "application/font-sfnt",
	".otf":    "application/font-sfnt",
	".dex":    "application/vnd.android.dex",
	".dey":    "application/vnd.android.dey",
}

const (
	// DefaultExpireTime default signed url expire time, 10min
	DefaultExpireTime = time.Second * 10 * 60
)

func init() {
	for key, val := range mimeTypes {
		mime.AddExtensionType(key, val)
	}
}

// TypeByExtension returns the MIME type associated with the file extension ext
// The extension ext should begin with a leading dot, as in ".html"
// When ext has no associated type, TypeByExtension returns "application/octet-stream"
func TypeByExtension(ext string) (mimeType string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	if mimeType = mime.TypeByExtension(ext); mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return
}
