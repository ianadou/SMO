// SMO — Primary button
// Full-width on mobile (parent decides). Background flips on hover; spinner replaces text on loading.

const PrimaryButton = ({ children, loading = false, disabled = false, type = 'submit', onClick }) => {
  const isDisabled = disabled || loading;
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={isDisabled}
      className={`smo-btn smo-btn-primary${isDisabled ? ' is-disabled' : ''}`}
    >
      {loading ? (
        <>
          <IconLoader size={16} className="smo-spin" />
          <span>Connexion…</span>
        </>
      ) : (
        children
      )}
    </button>
  );
};

Object.assign(window, { PrimaryButton });
