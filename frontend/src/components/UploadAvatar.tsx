import { useRef, useState, type ChangeEvent } from 'react';
import { uploadUserImage } from '../api/users';

interface UploadAvatarProps {
  src: string | null;
  name?: string;
  onUploaded(imageId: number, previewUrl: string): void;
  onError?(message: string): void;
}

export function UploadAvatar({ src, name, onUploaded, onError }: UploadAvatarProps) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [busy, setBusy] = useState(false);

  function pick() {
    inputRef.current?.click();
  }

  async function onChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;

    setBusy(true);
    try {
      const { id } = await uploadUserImage(file);
      const reader = new FileReader();
      reader.onload = (ev) => {
        const result = typeof ev.target?.result === 'string' ? ev.target.result : '';
        onUploaded(id, result);
      };
      reader.readAsDataURL(file);
    } catch (err) {
      onError?.(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div>
      <div className="avatar-lg-wrap">
        <span
          className="avatar avatar-lg"
          onClick={pick}
          style={{ cursor: 'pointer', borderRadius: 5, opacity: busy ? 0.6 : 1 }}
          title={name}
        >
          {src ? (
            <img src={src} alt={name ?? ''} />
          ) : (
            <div className="avatar-empty-stack">
              <span className="avatar-empty-arrow" aria-hidden="true">
                ↑
              </span>
              <span className="avatar-empty-label">
                загрузить
                <br />
                аватар
              </span>
            </div>
          )}
        </span>
      </div>
      {busy || src ? (
        <div className="avatar-change-hint">
          {busy ? 'загрузка…' : 'нажмите чтобы изменить'}
        </div>
      ) : null}
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        onChange={onChange}
        style={{ display: 'none' }}
      />
    </div>
  );
}
