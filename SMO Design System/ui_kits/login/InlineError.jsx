// SMO — InlineError
// Renders below the form. Never as a toast.

const InlineError = ({ children }) => (
  <div className="smo-error" role="alert">
    <IconAlertCircle size={16} />
    <span>{children}</span>
  </div>
);

Object.assign(window, { InlineError });
