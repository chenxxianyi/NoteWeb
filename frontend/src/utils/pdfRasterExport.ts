export interface RasterPdfPage {
  width: number
  height: number
  jpegData: Uint8Array
}

function asciiBytes(value: string): Uint8Array {
  const bytes = new Uint8Array(value.length)
  for (let index = 0; index < value.length; index++) {
    bytes[index] = value.charCodeAt(index) & 0xff
  }
  return bytes
}

function concatBytes(parts: Uint8Array[]): Uint8Array<ArrayBuffer> {
  const length = parts.reduce((sum, part) => sum + part.length, 0)
  const output = new Uint8Array(new ArrayBuffer(length))
  let offset = 0
  for (const part of parts) {
    output.set(part, offset)
    offset += part.length
  }
  return output
}

function formatNumber(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(3).replace(/\.?0+$/, '')
}

export function createRasterPdf(pages: RasterPdfPage[]): Blob {
  if (pages.length === 0) {
    throw new Error('No pages to export')
  }

  const chunks: Uint8Array[] = [asciiBytes('%PDF-1.4\n%\xE2\xE3\xCF\xD3\n')]
  const offsets = [0]
  let byteOffset = chunks[0].length

  const push = (chunk: Uint8Array) => {
    chunks.push(chunk)
    byteOffset += chunk.length
  }

  const writeObject = (id: number, bodyParts: Uint8Array[]) => {
    offsets[id] = byteOffset
    push(asciiBytes(`${id} 0 obj\n`))
    bodyParts.forEach(push)
    push(asciiBytes('\nendobj\n'))
  }

  const pageObjectIds = pages.map((_, index) => 3 + index * 3)
  const kids = pageObjectIds.map((id) => `${id} 0 R`).join(' ')

  writeObject(1, [asciiBytes('<< /Type /Catalog /Pages 2 0 R >>')])
  writeObject(2, [asciiBytes(`<< /Type /Pages /Kids [${kids}] /Count ${pages.length} >>`)])

  pages.forEach((page, index) => {
    const pageId = 3 + index * 3
    const contentId = pageId + 1
    const imageId = pageId + 2
    const width = Math.max(1, Math.round(page.width))
    const height = Math.max(1, Math.round(page.height))
    const pdfWidth = formatNumber(width)
    const pdfHeight = formatNumber(height)
    const content = `q\n${pdfWidth} 0 0 ${pdfHeight} 0 0 cm\n/Im${index + 1} Do\nQ\n`

    writeObject(pageId, [
      asciiBytes(
        `<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${pdfWidth} ${pdfHeight}] ` +
        `/Resources << /XObject << /Im${index + 1} ${imageId} 0 R >> >> ` +
        `/Contents ${contentId} 0 R >>`,
      ),
    ])
    writeObject(contentId, [
      asciiBytes(`<< /Length ${asciiBytes(content).length} >>\nstream\n${content}endstream`),
    ])
    writeObject(imageId, [
      asciiBytes(
        `<< /Type /XObject /Subtype /Image /Width ${width} /Height ${height} ` +
        '/ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode ' +
        `/Length ${page.jpegData.length} >>\nstream\n`,
      ),
      page.jpegData,
      asciiBytes('\nendstream'),
    ])
  })

  const xrefOffset = byteOffset
  const objectCount = 3 + pages.length * 3
  push(asciiBytes(`xref\n0 ${objectCount}\n`))
  push(asciiBytes('0000000000 65535 f \n'))
  for (let id = 1; id < objectCount; id++) {
    push(asciiBytes(`${String(offsets[id]).padStart(10, '0')} 00000 n \n`))
  }
  push(asciiBytes(
    `trailer\n<< /Size ${objectCount} /Root 1 0 R >>\nstartxref\n${xrefOffset}\n%%EOF\n`,
  ))

  return new Blob([concatBytes(chunks)], { type: 'application/pdf' })
}
