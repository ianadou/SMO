// SMO — Invite page components
// Three views: A (initial), B (modal), C (response confirmed).
// Wordmark is loaded from ../login/Wordmark.jsx via the host HTML.

const MATCH = {
  organizer: "Alex L.",
  groupName: "Foot du jeudi",
  date: "Jeudi 7 mai 2026",
  time: "19h30",
  venue: "Salle Pierre Mendès, 2 rue des Sports, Lyon",
  capacity: "10 joueurs (5 vs 5)",
  confirmed: 6,
  total: 10,
  confirmedPlayers: [
    { name: "Inès R.",  initials: "IR" },
    { name: "Théo B.",  initials: "TB" },
    { name: "Marc R.",  initials: "MR" },
    { name: "Paul S.",  initials: "PS" },
    { name: "Yann N.",  initials: "YN" },
    { name: "Cédric D.", initials: "CD" },
  ],
};

// ---- Building blocks ---------------------------------------------------

const InvHeader = () => (
  <header className="inv-header">
    <Wordmark size={32} />
  </header>
);

const InvDetailsCard = ({ compact = false }) => (
  <div className={`inv-card${compact ? ' is-compact' : ''}`}>
    <div className="inv-row">
      <span className="inv-row-icon"><IconCalendar size={20} /></span>
      <span className="inv-row-text">{MATCH.date}</span>
    </div>
    <div className="inv-row">
      <span className="inv-row-icon"><IconClock size={20} /></span>
      <span className="inv-row-text"><span className="num">{MATCH.time}</span></span>
    </div>
    <div className="inv-row">
      <span className="inv-row-icon"><IconMapPin size={20} /></span>
      <span className="inv-row-text">{MATCH.venue}</span>
    </div>
    <div className="inv-row">
      <span className="inv-row-icon"><IconUsers size={20} /></span>
      <span className="inv-row-text">
        <span className="num">{MATCH.capacity.split(' ')[0]}</span>
        <span> joueurs </span>
        <span className="muted">({MATCH.capacity.match(/\(([^)]+)\)/)?.[1]})</span>
      </span>
    </div>
  </div>
);

const InvConfirmedBlock = () => {
  const visible = MATCH.confirmedPlayers.slice(0, 6);
  return (
    <div className="inv-confirmed">
      <div className="inv-confirmed-head">
        <span>Déjà confirmés</span>
        <span><span className="num">{MATCH.confirmed}</span><span> / </span><span className="num">{MATCH.total}</span></span>
      </div>
      <div className="inv-avatars">
        {visible.map((p, i) => (
          <span key={i} className="inv-avatar" title={p.name}>{p.initials}</span>
        ))}
      </div>
    </div>
  );
};

const InvLegal = () => (
  <p className="inv-legal">
    SMO ne stocke ni votre numéro ni votre email.<br />
    Votre réponse est liée au lien d'invitation reçu.
  </p>
);

const InvLegalAfter = () => (
  <p className="inv-legal">
    Vous pouvez fermer cette page.<br />
    Votre réponse est sauvegardée.
  </p>
);

// ---- Modal -------------------------------------------------------------

const InvModal = ({ onYes, onNo, onCancel }) => {
  const onOverlayClick = (e) => {
    if (e.target === e.currentTarget) onCancel();
  };
  return (
    <div className="inv-modal-overlay" onClick={onOverlayClick} role="dialog" aria-modal="true">
      <div className="inv-modal">
        <h2 className="inv-modal-title">Vous venez à ce match ?</h2>
        <div className="inv-modal-actions">
          <button className="inv-btn inv-btn-yes" onClick={onYes}>
            <IconCheck size={18} />
            <span>Oui, je viens</span>
          </button>
          <button className="inv-btn inv-btn-no" onClick={onNo}>
            <IconX size={18} />
            <span>Non, je ne peux pas</span>
          </button>
        </div>
        <button className="inv-btn-link inv-modal-cancel" onClick={onCancel}>Annuler</button>
      </div>
    </div>
  );
};

// ---- Vue A : Invitation initiale ---------------------------------------

const InvViewA = ({ onRespond }) => (
  <>
    <InvHeader />
    <section className="inv-hero">
      <h1 className="inv-hero-title">Vous êtes invité</h1>
      <p className="inv-hero-sub">
        <strong>{MATCH.organizer}</strong> vous invite au match du groupe <strong>{MATCH.groupName}</strong>
      </p>
    </section>
    <InvDetailsCard />
    <InvConfirmedBlock />
    <button className="inv-btn inv-btn-primary" onClick={onRespond}>
      <IconCheckCircle size={18} />
      <span>Répondre</span>
    </button>
    <InvLegal />
  </>
);

// ---- Vue C : Confirmation de réponse -----------------------------------

const InvViewC = ({ answer, onModify, locked = false }) => {
  const isYes = answer === 'yes';
  return (
    <>
      <InvHeader />
      <section className="inv-result-hero">
        <span className={`inv-result-icon ${isYes ? 'is-yes' : 'is-no'}`}>
          {isYes ? <IconCheckCircleFilled size={64} /> : <IconXCircleFilled size={64} />}
        </span>
        <h1 className="inv-result-title">
          {isYes ? "Vous êtes inscrit" : "Réponse enregistrée"}
        </h1>
        <p className="inv-result-sub">
          {isYes
            ? <>Rendez-vous <strong style={{color:'var(--fg-default)', fontWeight:500}}>{MATCH.date}</strong> à <span className="num" style={{color:'var(--fg-default)'}}>{MATCH.time}</span></>
            : "Vous ne participerez pas à ce match"}
        </p>
      </section>
      <InvDetailsCard compact />
      {locked ? (
        <div className="inv-locked">
          <span className="inv-row-icon"><IconUsers size={18} /></span>
          <span>Les équipes ont été formées, votre réponse ne peut plus être modifiée.</span>
        </div>
      ) : (
        <button className="inv-btn inv-btn-secondary" onClick={onModify}>
          Modifier ma réponse
        </button>
      )}
      <InvLegalAfter />
    </>
  );
};

// ---- Stateful screen wrapper ------------------------------------------

const InvScreen = ({ initialView = 'A', initialAnswer = null, locked = false, frozen = false }) => {
  // frozen: render exactly the initialView with no internal transitions.
  // Used in the gallery so each card stays on its labelled state.
  const [view, setView] = React.useState(initialView);
  const [answer, setAnswer] = React.useState(initialAnswer);
  const [modal, setModal] = React.useState(initialView === 'B');

  const handleRespond = () => { if (frozen) return; setModal(true); };
  const handleYes = () => {
    if (frozen) return;
    setAnswer('yes'); setModal(false); setView('C');
  };
  const handleNo = () => {
    if (frozen) return;
    setAnswer('no'); setModal(false); setView('C');
  };
  const handleCancel = () => { if (frozen) return; setModal(false); if (view === 'A' && !answer) setView('A'); };
  const handleModify = () => { if (frozen) return; setModal(true); };

  // For frozen variants, derive what to render directly.
  const renderView = frozen ? initialView : view;
  const renderAnswer = frozen ? initialAnswer : answer;
  const showModal = frozen ? (initialView === 'B') : modal;

  return (
    <div className="inv-screen">
      {(renderView === 'A' || (frozen && initialView === 'B')) && (
        <InvViewA onRespond={handleRespond} />
      )}
      {renderView === 'C' && (
        <InvViewC answer={renderAnswer} onModify={handleModify} locked={locked} />
      )}
      {showModal && (
        <InvModal
          onYes={handleYes}
          onNo={handleNo}
          onCancel={handleCancel}
        />
      )}
    </div>
  );
};

Object.assign(window, {
  MATCH, InvHeader, InvDetailsCard, InvConfirmedBlock,
  InvLegal, InvLegalAfter, InvModal,
  InvViewA, InvViewC, InvScreen,
});
