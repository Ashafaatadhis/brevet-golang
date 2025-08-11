package helpers

import (
	"strings"

	"baliance.com/gooxml/document"
)

// ReplaceAll placeholders in doc, lebih tahan run split
func ReplaceAll(doc *document.Document, placeholders map[string]string) {
	for paraIdx := 0; paraIdx < len(doc.Paragraphs()); paraIdx++ {
		p := doc.Paragraphs()[paraIdx]

		// gabung semua runs jadi satu string
		var fullText strings.Builder
		for _, run := range p.Runs() {
			fullText.WriteString(run.Text())
		}

		fullStr := fullText.String()

		// cek tiap placeholder, kalau ada, replace di seluruh paragraph runs
		for key, val := range placeholders {
			if strings.Contains(fullStr, key) {
				// hapus semua run text di paragraph ini
				for _, run := range p.Runs() {
					run.Clear()
				}
				// set ulang text gabungan yang sudah diganti
				newText := strings.ReplaceAll(fullStr, key, val)
				// buat satu run baru dengan teks baru
				p.AddRun().AddText(newText)
				// update fullStr supaya tidak replace berulang jika ada beberapa placeholder di satu paragraph
				fullStr = newText
			}
		}
	}
}
