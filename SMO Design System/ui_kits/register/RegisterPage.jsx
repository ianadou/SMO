// SMO — Register flow components
// Reuses TextField, PrimaryButton, InlineError, Wordmark from ../login/.
// Three screens: RegisterForm (with 4 states) → CheckEmail → Verifying (success / token-expired).

// ---- Password strength heuristic --------------------------------------
// Simple visual feedback. The product spec is "no rules text" — just a 4-segment bar.
// Levels:
//   0: empty
//   1: <8 chars
//   2: ≥8 chars, single character class
//   3: ≥8 chars + 2 classes (or ≥10 chars)
//   4: ≥10 chars + 3+ classes
function passwordStrength(pwd) {
  if (!pwd) return 0;
  const classes =
    /[a-z]/.test(pwd) +
    /[A-Z]/.test(pwd) +
    /[0-9]/.test(pwd) +
    /[^A-Za-z0-9]/.test(pwd);
  if (pwd.length < 8) return 1;
  if (pwd.length >= 12 && classes >= 3) return 4;
  if (classes >= 2 && pwd.length >= 8) return 3;
  return 2;
}
const STRENGTH_LABELS = ['', 'Faible', 'Moyen', 'Fort', 'Très fort'];

const StrengthMeter = ({ level }) => (
  <div>
    <div className="reg-strength" aria-hidden="true">
      {[1,2,3,4].map(i => (
        <span
          key={i}
          className={`reg-strength-seg${i <= level ? ` is-${level}` : ''}`}
        />
      ))}
    </div>
    <div className="reg-strength-label">
      <span>Au moins 8 caractères</span>
      <span className="num">{STRENGTH_LABELS[level] || ''}</span>
    </div>
  </div>
);

// ---- Custom checkbox --------------------------------------------------

const CguCheckbox = ({ checked, onChange, hasError = false }) => (
  <label className="reg-check">
    <input type="checkbox" checked={checked} onChange={(e) => onChange?.(e.target.checked)} aria-invalid={hasError ? 'true' : 'false'}/>
    <span className="reg-check-box" style={hasError ? { boxShadow: 'inset 0 0 0 1px #DC2A3B' } : undefined}>
      {checked && <IconCheck size={14}/>}
    </span>
    <span className="reg-check-label">
      J'accepte les <a href="#cgu" target="_blank" rel="noopener noreferrer">conditions générales</a>
      {' '}et la <a href="#privacy" target="_blank" rel="noopener noreferrer">politique de confidentialité</a>
    </span>
  </label>
);

// Small inline check icon for the checkbox tick. Reuses lucide proportions.
const IconCheck = ({ size = 14 }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24"
    fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="20 6 9 17 4 12"/>
  </svg>
);

// ---- Screen 1 — Register form -----------------------------------------

const RegisterForm = ({
  // Allow controlled values for the gallery preset states.
  initial = {},
  forcedError = null,
  forcedLoading = false,
  forcedFocus = null,
  frozen = false,
}) => {
  const [name, setName]   = React.useState(initial.name   ?? '');
  const [email, setEmail] = React.useState(initial.email  ?? '');
  const [pwd, setPwd]     = React.useState(initial.pwd    ?? '');
  const [pwdShown, setPwdShown] = React.useState(false);
  const [cgu, setCgu]     = React.useState(initial.cgu    ?? false);
  const [submitted, setSubmitted] = React.useState(false);

  const strength = passwordStrength(pwd);

  // Validations
  const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  const nameValid = name.trim().length >= 2 && name.trim().length <= 50;
  const emailValidFormat = emailRe.test(email);
  const pwdValid = pwd.length >= 8;
  const allValid = nameValid && emailValidFormat && pwdValid && cgu;

  // Forced errors override real-time validity (so we can render the gallery state)
  const isEmailFormatError = forcedError === 'email-format' || (submitted && email && !emailValidFormat);
  const isEmailExistsError = forcedError === 'email-exists';
  const isCguError = forcedError === 'cgu-missing' || (submitted && !cgu);

  const handleSubmit = (e) => {
    e?.preventDefault?.();
    if (frozen) return;
    setSubmitted(true);
    if (!allValid) return;
    // In a real app: POST /register here.
  };

  return (
    <form className="reg-form" onSubmit={handleSubmit} noValidate>
      <div className="reg-mark"><Wordmark size={32}/></div>
      <h1 className="reg-title">Créer un compte</h1>
      <p className="reg-sub">Pour organiser vos matchs et inviter vos joueurs.</p>

      <div className="reg-stack">
        {/* Name */}
        <div>
          <TextField
            id="reg-name"
            label="Comment vous présenter à vos joueurs ?"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Alex L."
            autoComplete="name"
            autoFocus={forcedFocus === 'name'}
          />
          <p className="reg-hint">C'est ce que verront vos joueurs dans les invitations (ex : « Alex L. vous invite… »).</p>
        </div>

        {/* Email */}
        <div>
          <TextField
            id="reg-email"
            label="Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="alex@exemple.fr"
            autoComplete="email"
            inputMode="email"
            hasError={isEmailFormatError || isEmailExistsError}
            autoFocus={forcedFocus === 'email'}
          />
          {isEmailFormatError && <p className="reg-hint is-error">Format d'email invalide.</p>}
          {isEmailExistsError && <p className="reg-hint is-error">Un compte existe déjà pour cette adresse. <a href="#login" style={{color:'var(--action-primary)'}}>Se connecter</a></p>}
        </div>

        {/* Password */}
        <div>
          <TextField
            id="reg-pwd"
            label="Mot de passe"
            type={pwdShown ? 'text' : 'password'}
            value={pwd}
            onChange={(e) => setPwd(e.target.value)}
            placeholder="••••••••"
            autoComplete="new-password"
            autoFocus={forcedFocus === 'pwd'}
            rightSlot={
              <button
                type="button"
                className="smo-icon-btn"
                aria-label={pwdShown ? 'Cacher le mot de passe' : 'Afficher le mot de passe'}
                onClick={() => setPwdShown(s => !s)}
                tabIndex={0}
              >
                {pwdShown ? <IconEyeOff size={18}/> : <IconEye size={18}/>}
              </button>
            }
          />
          <StrengthMeter level={strength}/>
        </div>

        {/* CGU */}
        <div>
          <CguCheckbox checked={cgu} onChange={setCgu} hasError={isCguError}/>
          {isCguError && <p className="reg-hint is-error" style={{marginTop:4}}>Vous devez accepter les conditions pour créer un compte.</p>}
        </div>
      </div>

      <PrimaryButton
        type="submit"
        loading={forcedLoading}
        disabled={!forcedLoading && !allValid}
      >
        {forcedLoading ? null : (allValid ? 'Créer mon compte' : 'Compléter le formulaire')}
      </PrimaryButton>

      <p className="smo-secondary">
        Déjà un compte ? <a href="#login" className="smo-link">Se connecter</a>
      </p>
    </form>
  );
};

// ---- Screen 2 — Check email -------------------------------------------

const CheckEmailScreen = ({ email = 'alex@exemple.fr', initialCountdown = 60, onBack, frozen = false }) => {
  const [countdown, setCountdown] = React.useState(initialCountdown);

  React.useEffect(() => {
    if (frozen) return;
    if (countdown <= 0) return;
    const t = setTimeout(() => setCountdown(c => c - 1), 1000);
    return () => clearTimeout(t);
  }, [countdown, frozen]);

  const canResend = countdown <= 0;
  const handleResend = () => { if (frozen) return; if (canResend) setCountdown(60); };

  return (
    <div className="reg-confirm">
      <div className="reg-mark"><Wordmark size={32}/></div>

      <div className="reg-hero">
        <span className="reg-hero-icon is-primary"><IconMailCheck size={64}/></span>
        <h1 className="reg-hero-title">Vérifiez votre boîte mail</h1>
        <p className="reg-hero-sub">Nous avons envoyé un lien de connexion à</p>
        <div className="reg-email-value">{email}</div>
      </div>

      <ol className="reg-steps">
        <li className="reg-step">
          <span className="reg-step-num">1</span>
          <span className="reg-step-text">Ouvrez votre boîte mail (vérifiez aussi les spams).</span>
        </li>
        <li className="reg-step">
          <span className="reg-step-num">2</span>
          <span className="reg-step-text">Cliquez sur le lien de connexion.</span>
        </li>
        <li className="reg-step">
          <span className="reg-step-num">3</span>
          <span className="reg-step-text">Vous serez automatiquement connecté.</span>
        </li>
      </ol>

      <button
        type="button"
        className="reg-resend"
        disabled={!canResend}
        onClick={handleResend}
      >
        {canResend
          ? "Renvoyer l'email"
          : <>Renvoyer l'email (<span className="num">{countdown}s</span>)</>}
      </button>

      <p className="reg-tertiary">
        Mauvaise adresse ? <a href="#" onClick={(e) => { e.preventDefault(); onBack?.(); }}>Modifier mon email</a>
      </p>
    </div>
  );
};

// ---- Screen 3 — Verifying ---------------------------------------------

const VerifyingScreen = ({ status = 'success' }) => {
  if (status === 'success') {
    return (
      <div className="reg-confirm">
        <div className="reg-mark"><Wordmark size={32}/></div>
        <div className="reg-hero">
          <span className="reg-hero-icon is-success"><IconCheckCircleFilledReg size={80}/></span>
          <h1 className="reg-hero-title">Email vérifié !</h1>
          <p className="reg-hero-sub">Connexion en cours…</p>
          <span className="reg-spin-wrap"><IconLoader size={20} className="smo-spin"/></span>
        </div>
      </div>
    );
  }
  // Error variant
  return (
    <div className="reg-confirm">
      <div className="reg-mark"><Wordmark size={32}/></div>
      <div className="reg-hero">
        <span className="reg-hero-icon is-error"><IconXCircleFilledReg size={80}/></span>
        <h1 className="reg-hero-title">Lien expiré ou invalide</h1>
        <p className="reg-hero-sub">Le lien que vous avez utilisé n'est plus valide. Recréez votre compte ou demandez un nouveau lien.</p>
      </div>
      <a href="#login" className="smo-btn smo-btn-primary" style={{textDecoration:'none'}}>
        Retour à la connexion
      </a>
    </div>
  );
};

Object.assign(window, {
  passwordStrength, StrengthMeter, CguCheckbox,
  RegisterForm, CheckEmailScreen, VerifyingScreen,
});
