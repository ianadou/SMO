// SMO — LoginForm
// The whole login surface: wordmark + title + form + secondary link + footer note.
// Stateful: holds email, password, password-visibility, submission state, error.
// `forcedState` lets the static gallery render any state without user interaction.

const { useState, useRef, useEffect } = React;

const LoginForm = ({ forcedState = null, compact = false }) => {
  const [email, setEmail] = useState(
    forcedState === 'error' || forcedState === 'loading' || forcedState === 'focus'
      ? 'alex@gmail.com' : ''
  );
  const [password, setPassword] = useState(
    forcedState === 'error' || forcedState === 'loading' ? '••••••••' : ''
  );
  const [showPwd, setShowPwd] = useState(false);
  const [loading, setLoading] = useState(forcedState === 'loading');
  const [error, setError] = useState(forcedState === 'error' ? 'Identifiants incorrects.' : '');

  const emailRef = useRef(null);
  const passRef = useRef(null);

  // Focus initial — only on the live (interactive) instance, not the static gallery.
  useEffect(() => {
    if (forcedState === null && emailRef.current) {
      // Non-aggressive — only when the document hasn't focused something else.
      if (document.activeElement === document.body) {
        emailRef.current.focus();
      }
    }
    if (forcedState === 'focus' && passRef.current) {
      passRef.current.focus();
    }
  }, [forcedState]);

  const submit = (e) => {
    if (e) e.preventDefault();
    if (forcedState !== null) return; // gallery is non-interactive
    if (!email || !password) return;
    setError('');
    setLoading(true);
    // Fake a request; resolve as error to demo the error state on real form too.
    setTimeout(() => {
      setLoading(false);
      setError('Identifiants incorrects.');
    }, 1400);
  };

  const onPasswordKeyDown = (e) => {
    if (e.key === 'Enter') submit(e);
  };

  const emailHasError = !!error;
  const passwordHasError = !!error;

  return (
    <form
      className={`smo-login${compact ? ' is-compact' : ''}`}
      onSubmit={submit}
      noValidate
    >
      <div className="smo-login-mark">
        <Wordmark size={compact ? 32 : 40} />
      </div>

      <h1 className="smo-title">Connexion organisateur</h1>

      <div className="smo-stack">
        <TextField
          id={`email-${forcedState ?? 'live'}`}
          label="Email"
          type="email"
          inputMode="email"
          autoComplete="email"
          placeholder="toi@exemple.fr"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          autoFocus={forcedState === null}
          hasError={emailHasError}
          inputRef={emailRef}
        />

        <TextField
          id={`password-${forcedState ?? 'live'}`}
          label="Mot de passe"
          type={showPwd ? 'text' : 'password'}
          autoComplete="current-password"
          placeholder="••••••••"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          hasError={passwordHasError}
          onKeyDown={onPasswordKeyDown}
          inputRef={passRef}
          rightSlot={
            <button
              type="button"
              className="smo-icon-btn"
              aria-label={showPwd ? 'Masquer le mot de passe' : 'Afficher le mot de passe'}
              onClick={() => setShowPwd((v) => !v)}
              tabIndex={0}
            >
              {showPwd ? <IconEyeOff size={18} /> : <IconEye size={18} />}
            </button>
          }
        />
      </div>

      <PrimaryButton loading={loading}>Se connecter</PrimaryButton>

      {error && <InlineError>{error}</InlineError>}

      <div className="smo-secondary">
        Pas encore de compte ?{' '}
        <a href="#" className="smo-link">S'inscrire</a>
      </div>

      <div className="smo-footnote">
        Les joueurs n'ont pas besoin de compte — ils accèdent par lien d'invitation.
      </div>
    </form>
  );
};

Object.assign(window, { LoginForm });
