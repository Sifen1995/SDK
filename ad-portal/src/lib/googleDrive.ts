const DRIVE_FILE_PATTERNS = [
  /\/file\/d\/([a-zA-Z0-9_-]+)/,
  /[?&]id=([a-zA-Z0-9_-]+)/,
  /\/d\/([a-zA-Z0-9_-]+)/,
];

/** Extract Google Drive file ID from common share URL formats. */
export function extractGoogleDriveFileId(input: string): string | null {
  const trimmed = input.trim();
  if (!trimmed) return null;

  for (const pattern of DRIVE_FILE_PATTERNS) {
    const match = trimmed.match(pattern);
    if (match?.[1]) return match[1];
  }

  if (/^[a-zA-Z0-9_-]{10,}$/.test(trimmed)) {
    return trimmed;
  }

  return null;
}

/** Convert a Drive share link or file ID into a direct image URL for image_url. */
export function googleDriveToDirectImageUrl(input: string): string | null {
  const fileId = extractGoogleDriveFileId(input);
  if (!fileId) return null;
  return `https://drive.google.com/uc?export=view&id=${fileId}`;
}

export function isLikelyGoogleDriveUrl(input: string): boolean {
  return input.includes('drive.google.com') || extractGoogleDriveFileId(input) !== null;
}

export function isValidHttpUrl(input: string): boolean {
  try {
    const url = new URL(input);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}
