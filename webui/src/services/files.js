export const maxFileSize = 5 * 1024 * 1024;

const mimeTypePattern = /^[a-zA-Z0-9.+-]+\/[a-zA-Z0-9.+-]+$/;

export function readAsDataURL(file) {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.onload = () => resolve(reader.result);
		reader.onerror = () => reject(reader.error);
		reader.readAsDataURL(file);
	});
}

export function checkFile(file, imageOnly) {
	if (file.size === 0) {
		return "Il file scelto è vuoto.";
	}
	if (file.size > maxFileSize) {
		return "Il file supera il limite di 5 MB.";
	}
	if (imageOnly && !file.type.startsWith("image/")) {
		return "Il file scelto non è un'immagine.";
	}
	if (!imageOnly && [...file.name].length > 255) {
		return "Il nome del file è troppo lungo.";
	}
	return null;
}

export async function fileToAttachment(file) {
	const dataURL = await readAsDataURL(file);
	return {
		name: file.name,
		mimeType: mimeTypePattern.test(file.type) ? file.type : "application/octet-stream",
		data: dataURL.slice(dataURL.indexOf(",") + 1),
	};
}

export function attachmentURL(attachment) {
	return `data:${attachment.mimeType};base64,${attachment.data}`;
}
