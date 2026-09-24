import { useRef, useState, type ChangeEvent } from 'react';
import { uploadEventImage } from '../api/events';

interface UploadEventCoverProps {
  src: string | null;
  editable: boolean;
  // For new (unsaved) events we can't upload yet — we just stash the file
  // and base64 preview. The parent will perform the upload after the event
  // is created.
  eventId?: number | null;
  onPicked(file: File, previewUrl: string): void;
  onUploaded?(imageId: number, previewUrl: string): void;
  onError?(message: string): void;
}

export function UploadEventCover({
  src,
  editable,
  eventId,
  onPicked,
  onUploaded,
  onError,
}: UploadEventCoverProps) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [busy, setBusy] = useState(false);

  function pick() {
    if (!editable) return;
    inputRef.current?.click();
  }

  async function onChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (ev) => {
      const previewUrl =
        typeof ev.target?.result === 'string' ? ev.target.result : '';
      onPicked(file, previewUrl);
    };
    reader.readAsDataURL(file);

    if (eventId == null) return;

    setBusy(true);
    try {
      const { id } = await uploadEventImage(eventId, file);
      const previewReader = new FileReader();
      previewReader.onload = (ev) => {
        const result =
          typeof ev.target?.result === 'string' ? ev.target.result : '';
        onUploaded?.(id, result);
      };
      previewReader.readAsDataURL(file);
    } catch (err) {
      onError?.(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div
      className="avatar-event-lg"
      onClick={pick}
      style={{ cursor: editable ? 'pointer' : 'default', opacity: busy ? 0.6 : 1 }}
    >
      {src ? (
        <img src={src} alt="" />
      ) : (
        <span style={{ color: 'var(--fg-30)' }}>
          {editable ? '+ загрузить обложку' : 'без обложки'}
        </span>
      )}
      {editable && (
        <input
          ref={inputRef}
          type="file"
          accept="image/*"
          onChange={onChange}
          style={{ display: 'none' }}
        />
      )}
    </div>
  );
}
