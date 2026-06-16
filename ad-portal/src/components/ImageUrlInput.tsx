import { useEffect, useState } from 'react';
import {
  googleDriveToDirectImageUrl,
  isLikelyGoogleDriveUrl,
  isValidHttpUrl,
} from '../lib/googleDrive';

type ImageSource = 'url' | 'drive';

interface ImageUrlInputProps {
  value: string;
  onChange: (directUrl: string) => void;
  required?: boolean;
  label?: string;
}

declare global {
  interface Window {
    gapi: any;
    google: any;
  }
}

const CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID || '';
const API_KEY = import.meta.env.VITE_GOOGLE_API_KEY || '';
const APP_ID = CLIENT_ID ? CLIENT_ID.split('-')[0] : '';

export default function ImageUrlInput({
  value,
  onChange,
  required,
  label = 'Campaign image',
}: ImageUrlInputProps) {
  const [source, setSource] = useState<ImageSource>(() =>
    isLikelyGoogleDriveUrl(value) ? 'drive' : 'url',
  );
  const [urlInput, setUrlInput] = useState(value);
  const [previewError, setPreviewError] = useState('');
  
  const [pickerInited, setPickerInited] = useState(false);
  const [tokenClient, setTokenClient] = useState<any>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);

  useEffect(() => {
    if (source === 'url') {
      setUrlInput(value);
    }
  }, [value, source]);

  useEffect(() => {
    let loadGapi: any;
    if (source === 'drive') {
      loadGapi = setInterval(() => {
        if (window.gapi && window.google) {
          clearInterval(loadGapi);
          if (!pickerInited) {
            window.gapi.load('picker', { callback: () => setPickerInited(true) });
          }
          if (!tokenClient) {
            const client = window.google.accounts.oauth2.initTokenClient({
              client_id: CLIENT_ID,
              scope: 'https://www.googleapis.com/auth/drive.readonly',
              callback: (response: any) => {
                if (response.error !== undefined) {
                  setPreviewError('Authentication failed: ' + response.error);
                  return;
                }
                setAccessToken(response.access_token);
                createPicker(response.access_token);
              },
            });
            setTokenClient(client);
          }
        }
      }, 100);
    }
    return () => clearInterval(loadGapi);
  }, [source, pickerInited, tokenClient]);

  function applyUrl(raw: string) {
    setUrlInput(raw);
    setPreviewError('');
    if (!raw.trim()) {
      onChange('');
      return;
    }
    if (!isValidHttpUrl(raw)) {
      setPreviewError('Enter a valid http(s) URL');
      onChange('');
      return;
    }
    onChange(raw.trim());
  }

  function handleAuthClick() {
    setPreviewError('');
    if (!CLIENT_ID || !API_KEY) {
      setPreviewError('Google API credentials missing. Check your .env file.');
      return;
    }
    if (accessToken) {
      createPicker(accessToken);
    } else {
      if (!tokenClient) {
        setPreviewError('Google API scripts loading... please try again in a moment.');
        return;
      }
      tokenClient.requestAccessToken({ prompt: 'consent' });
    }
  }

  function createPicker(token: string) {
    if (!pickerInited) {
      setPreviewError('Google Picker API is still initializing.');
      return;
    }
    const view = new window.gapi.picker.View(window.gapi.picker.ViewId.DOCS);
    view.setMimeTypes('image/png,image/jpeg,image/jpg,image/webp,image/gif');

    const picker = new window.gapi.picker.PickerBuilder()
      .enableFeature(window.gapi.picker.Feature.NAV_HIDDEN)
      .setDeveloperKey(API_KEY)
      .setAppId(APP_ID)
      .setOAuthToken(token)
      .addView(view)
      .setCallback(pickerCallback)
      .build();
    picker.setVisible(true);
  }

  function pickerCallback(data: any) {
    if (data.action === window.gapi.picker.Action.PICKED) {
      const doc = data.docs[0];
      const mimeType = doc.mimeType;
      
      if (!mimeType.startsWith('image/')) {
        setPreviewError(`Please select a valid image file. You selected a ${mimeType} file.`);
        onChange('');
        return;
      }

      const fileId = doc.id;
      const direct = googleDriveToDirectImageUrl(fileId);
      if (direct) {
        onChange(direct);
        setPreviewError('');
      } else {
        setPreviewError('Failed to format Google Drive image URL.');
      }
    }
  }

  const previewUrl = value || undefined;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <label className="block text-sm font-medium text-primary">
          {label}
          {required && <span className="text-brand-600 ml-0.5">*</span>}
        </label>
        <div className="inline-flex rounded-lg bg-[var(--bg-subtle)] border border-[var(--border)] p-1 text-xs font-medium">
          <button
            type="button"
            onClick={() => setSource('url')}
            className={`rounded-md px-3 py-1.5 transition cursor-pointer ${
              source === 'url' ? 'bg-[var(--surface)] text-primary shadow-sm ring-1 ring-[var(--border)]' : 'text-muted hover:text-primary'
            }`}
          >
            Direct URL
          </button>
          <button
            type="button"
            onClick={() => setSource('drive')}
            className={`rounded-md px-3 py-1.5 transition cursor-pointer ${
              source === 'drive' ? 'bg-[var(--surface)] text-primary shadow-sm ring-1 ring-[var(--border)]' : 'text-muted hover:text-primary'
            }`}
          >
            Google Drive
          </button>
        </div>
      </div>

      {source === 'url' ? (
        <div className="space-y-2">
          <input
            type="url"
            value={urlInput}
            onChange={e => applyUrl(e.target.value)}
            placeholder="https://cdn.example.com/banner.png"
            className="field-input"
          />
          <p className="text-xs text-muted">Paste a publicly accessible image URL (CDN, S3, etc.).</p>
        </div>
      ) : (
        <div className="space-y-4">
          <button
            type="button"
            onClick={handleAuthClick}
            className="btn-secondary w-full py-3 border-dashed border-2 hover:border-[var(--color-brand-500)] dark:hover:bg-[var(--color-brand-900)] transition"
          >
            <span className="text-xl mr-2">📁</span>
            Select Image from Google Drive
          </button>
          <p className="text-xs text-muted text-center">
            Securely browse and select an image directly from your Google Drive.
          </p>
          {value && isLikelyGoogleDriveUrl(value) && (
            <p className="text-xs text-[var(--color-brand-600)] dark:text-[var(--color-brand-400)] break-all text-center">
              Selected: {value}
            </p>
          )}
        </div>
      )}

      {previewError && (
        <div className="alert-error">
          <p className="text-sm font-medium">{previewError}</p>
        </div>
      )}

      {previewUrl && !previewError && (
        <div className="rounded-xl border border-[var(--border)] bg-[var(--bg-subtle)] p-4">
          <p className="text-xs font-medium text-muted mb-2">Preview</p>
          <img
            src={previewUrl}
            alt="Campaign preview"
            className="max-h-40 rounded-lg border border-[var(--border)] object-contain bg-[var(--surface)] mx-auto"
            onError={() => setPreviewError('Image failed to load. Check sharing permissions or URL.')}
          />
        </div>
      )}
    </div>
  );
}
