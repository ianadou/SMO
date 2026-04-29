// SMO — TextField
// Floating label is intentionally NOT used. Static label above input, sober.
// Input: no border by default; bg-elevated fill against bg-base page.
// Focus: 2px inset ring in --action-primary. Error: 2px inset ring in --danger.

const TextField = ({
  id,
  label,
  type = 'text',
  value,
  onChange,
  placeholder,
  autoFocus = false,
  hasError = false,
  rightSlot = null,
  inputMode,
  autoComplete,
  onKeyDown,
  inputRef,
}) => {
  const ringStyle = hasError
    ? { boxShadow: 'inset 0 0 0 2px #DC2A3B' }
    : {};

  return (
    <div className="smo-field">
      <label htmlFor={id} className="smo-label">{label}</label>
      <div className="smo-input-wrap" style={ringStyle}>
        <input
          ref={inputRef}
          id={id}
          type={type}
          value={value}
          onChange={onChange}
          placeholder={placeholder}
          autoFocus={autoFocus}
          autoComplete={autoComplete}
          inputMode={inputMode}
          onKeyDown={onKeyDown}
          className="smo-input"
          aria-invalid={hasError ? 'true' : 'false'}
        />
        {rightSlot}
      </div>
    </div>
  );
};

Object.assign(window, { TextField });
