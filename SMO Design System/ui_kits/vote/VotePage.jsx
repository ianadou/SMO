// SMO — Vote post-match page
// Reuses: Wordmark (../login/Wordmark.jsx), MatchCardVs + Jersey + AvatarRing (../group_detail/...).

const VOTE_DATA = {
  match: {
    groupName: "Foot du jeudi",
    dateLabel: "Jeu 7 mai",
    winner: "red",  // équipe gagnante
  },
  voter: {
    name: "Alex L.",
    initials: "AL",
    team: "red",  // équipe rouge
    selfScore: 4.4,
    selfVotes: 4,
  },
  // 4 coéquipiers à noter (équipe rouge, sans le votant)
  teammates: [
    { initials: "IR", name: "Inès R.",  matchesTogether: 28, finalScore: 4.4, votes: 4, delta: +0.2 },
    { initials: "TB", name: "Théo B.",  matchesTogether: 18, finalScore: 3.8, votes: 4, delta: -0.1 },
    { initials: "MR", name: "Marc R.",  matchesTogether: 22, finalScore: 4.0, votes: 4, delta: +0.1 },
    { initials: "PS", name: "Paul S.",  matchesTogether: 15, finalScore: 3.6, votes: 4, delta: -0.2 },
  ],
  status: { voted: 8, total: 10 },
};

// ---- Building blocks ---------------------------------------------------

const VHeader = () => (
  <header className="vote-header"><Wordmark size={32}/></header>
);

const TeamBanner = ({ team }) => {
  const label = team === "red" ? "rouge" : "verte";
  return (
    <div className="vote-team-banner" role="note">
      <span className={`vote-team-dot is-${team}`} aria-hidden="true"/>
      <span>Vous étiez dans l'<strong style={{fontWeight:500}}>Équipe {label}</strong></span>
    </div>
  );
};

const StarRow = ({ value, onChange, readonly = false }) => {
  const stars = [1,2,3,4,5];
  return (
    <div className="vote-stars" role="radiogroup" aria-label="Note de 1 à 5 étoiles">
      {stars.map(n => {
        const active = n <= value;
        return (
          <button
            key={n}
            type="button"
            className={`vote-star-btn${active ? ' is-active' : ''}`}
            onClick={() => !readonly && onChange?.(n === value ? 0 : n)}
            aria-label={`${n} étoile${n > 1 ? 's' : ''}`}
            aria-pressed={active}
            disabled={readonly}
          >
            {active ? <IconStarFilled size={20}/> : <IconStarOutline size={20}/>}
          </button>
        );
      })}
    </div>
  );
};

const TeammateRow = ({ player, value, onChange, readonly }) => (
  <div className="vote-prow">
    <AvatarRing initials={player.initials} size={44} score={0} />
    <div className="vote-prow-mid">
      <div className="vote-prow-name">{player.name}</div>
      <div className="vote-prow-meta">
        <span className="num">{player.matchesTogether}</span>
        <span> matchs joués ensemble</span>
      </div>
    </div>
    <StarRow value={value} onChange={onChange} readonly={readonly}/>
  </div>
);

// ---- Vue A : Vote initial ----------------------------------------------

const ViewA = ({ initialVotes, onSubmit, frozen = false }) => {
  const [votes, setVotes] = React.useState(initialVotes || {});
  const setVote = (idx, n) => {
    if (frozen) return;
    setVotes(v => ({ ...v, [idx]: n }));
  };
  const filled = Object.values(votes).filter(v => v > 0).length;
  const total = VOTE_DATA.teammates.length;
  const allDone = filled === total;
  const pct = (filled / total) * 100;

  return (
    <>
      <VHeader/>
      <div className="vote-mc-wrap">
        <MatchCardVs status="finished" dateLabel={VOTE_DATA.match.dateLabel} winner={VOTE_DATA.match.winner}/>
      </div>
      <TeamBanner team={VOTE_DATA.voter.team}/>

      <section className="vote-section">
        <h2 className="vote-section-title">Notez vos coéquipiers</h2>
        <p className="vote-section-sub">Donnez une note de 1 à 5 étoiles à chacun de vos 4 coéquipiers.</p>
        <div className="vote-list">
          {VOTE_DATA.teammates.map((p, i) => (
            <TeammateRow
              key={i}
              player={p}
              value={votes[i] || 0}
              onChange={(n) => setVote(i, n)}
              readonly={frozen}
            />
          ))}
        </div>
        <div className="vote-counter">
          <div className="vote-counter-row">
            <span><span className="num">{filled}</span><span> / </span><span className="num">{total}</span><span> notes</span></span>
          </div>
          <div className="vote-progress" aria-hidden="true">
            <div className="vote-progress-fill" style={{ width: `${pct}%` }}/>
          </div>
        </div>
      </section>

      <p className="vote-legal">
        Le vote est définitif et anonyme.<br/>
        Vos coéquipiers ne sauront pas qui les a notés.
      </p>

      <div className="vote-sticky">
        <div className="vote-sticky-inner">
          <button
            className="vote-btn vote-btn-primary"
            disabled={!allDone}
            onClick={() => allDone && onSubmit?.(votes)}
          >
            {allDone ? "Soumettre mes votes (définitif)" : "Notez les 4 coéquipiers"}
          </button>
        </div>
      </div>
    </>
  );
};

// ---- Vue B : Modal de confirmation -------------------------------------

const SummaryStars = ({ value }) => (
  <span className="vote-summary-stars" aria-label={`${value} étoiles`}>
    {[1,2,3,4,5].map(n => (
      n <= value
        ? <IconStarFilled key={n} size={14}/>
        : <IconStarOutline key={n} size={14} style={{color:'var(--fg-muted)'}}/>
    ))}
  </span>
);

const VoteModal = ({ votes, onConfirm, onCancel }) => {
  const onOverlayClick = (e) => { if (e.target === e.currentTarget) onCancel?.(); };
  return (
    <div className="vote-modal-overlay" onClick={onOverlayClick} role="dialog" aria-modal="true">
      <div className="vote-modal">
        <span className="vote-modal-icon"><IconAlertTriangle size={32}/></span>
        <h2 className="vote-modal-title">Confirmer vos votes ?</h2>
        <p className="vote-modal-sub">Cette action est définitive. Vous ne pourrez plus modifier vos notes.</p>
        <div className="vote-modal-summary">
          {VOTE_DATA.teammates.map((p, i) => (
            <div key={i} className="vote-summary-row">
              <AvatarRing initials={p.initials} size={28} score={0}/>
              <span className="vote-summary-name">{p.name}</span>
              <SummaryStars value={votes[i] || 0}/>
            </div>
          ))}
        </div>
        <div className="vote-modal-actions">
          <button className="vote-btn vote-btn-primary" onClick={onConfirm}>
            <IconCheck size={18}/>
            <span>Confirmer</span>
          </button>
          <button className="vote-btn-link" onClick={onCancel}>Modifier</button>
        </div>
      </div>
    </div>
  );
};

// ---- Vue C : Vote soumis -----------------------------------------------

const ViewC = () => {
  const { voted, total } = VOTE_DATA.status;
  const pct = (voted / total) * 100;
  return (
    <>
      <VHeader/>
      <section className="vote-result-hero">
        <span className="vote-result-icon"><IconCheckCircleFilled size={64}/></span>
        <h1 className="vote-result-title">Merci pour votre vote</h1>
        <p className="vote-result-sub">Votre vote est enregistré</p>
      </section>

      <div className="vote-status-card">
        <div className="vote-status-head">
          <span className="vote-status-title">Votes en cours</span>
          <span className="vote-status-count">
            <span className="num">{voted}</span><span className="muted"> / {total}</span>
          </span>
        </div>
        <div className="vote-progress" aria-hidden="true">
          <div className="vote-progress-fill" style={{ width: `${pct}%` }}/>
        </div>
        <p className="vote-status-sub">
          Les résultats seront disponibles dès que tous les joueurs auront voté
          ou que l'organisateur aura clôturé le match.
        </p>
      </div>

      <div className="vote-mc-wrap">
        <MatchCardVs status="finished" dateLabel={VOTE_DATA.match.dateLabel} winner={VOTE_DATA.match.winner}/>
      </div>

      <p className="vote-legal">
        Vous pouvez fermer cette page.<br/>
        Nous vous notifierons quand les résultats seront disponibles.
      </p>
    </>
  );
};

// ---- Vue D : Match clôturé, résultats ----------------------------------

const Delta = ({ value }) => {
  if (value == null) return null;
  const up = value > 0;
  const down = value < 0;
  if (!up && !down) return null;
  const sign = up ? '+' : '';
  return (
    <span className={`vote-delta ${up ? 'is-up' : 'is-down'}`}>
      {up ? <IconTrendingUp size={12}/> : <IconTrendingDown size={12}/>}
      <span>{sign}{value.toFixed(1)} vs précédent</span>
    </span>
  );
};

const ResultRow = ({ player }) => (
  <div className="vote-result-row">
    <AvatarRing initials={player.initials} size={48} score={player.finalScore}/>
    <div className="vote-result-mid">
      <div className="vote-result-name">{player.name}</div>
      <div className="vote-result-meta">
        <span className="num">{player.votes}</span>
        <span> votes reçus</span>
      </div>
      <Delta value={player.delta}/>
    </div>
    <div className="vote-result-score-num">
      <span>{player.finalScore.toFixed(1)}</span>
      <span className="muted">/5</span>
    </div>
  </div>
);

const ViewD = () => {
  const teamLabel = VOTE_DATA.voter.team === 'red' ? 'rouge' : 'verte';
  return (
    <>
      <VHeader/>
      <div className="vote-mc-wrap">
        <MatchCardVs status="closed" dateLabel={VOTE_DATA.match.dateLabel} winner={VOTE_DATA.match.winner}/>
      </div>
      <p className="vote-section-sub" style={{margin:'0 0 var(--space-5)'}}>Vos coéquipiers ont reçu vos notes.</p>

      <section className="vote-results-section">
        <h2 className="vote-results-title">Notes finales de l'équipe {teamLabel}</h2>
        <div className="vote-list">
          {VOTE_DATA.teammates.map((p, i) => <ResultRow key={i} player={p}/>)}
        </div>
      </section>

      <section className="vote-results-section">
        <h2 className="vote-results-title">Votre note moyenne ce match</h2>
        <div className="vote-self-card">
          <AvatarRing initials={VOTE_DATA.voter.initials} size={88} score={VOTE_DATA.voter.selfScore} stroke={4}/>
          <div className="vote-self-score-big">
            <span>{VOTE_DATA.voter.selfScore.toFixed(1)}</span>
            <span className="muted">/5</span>
          </div>
          <div className="vote-self-meta">
            <span className="num">{VOTE_DATA.voter.selfVotes}</span>
            <span> votes reçus</span>
          </div>
        </div>
      </section>

      <button className="vote-btn vote-btn-secondary" disabled title="Réservé aux organisateurs connectés">
        Voir tous les classements du groupe
        <IconArrowRight size={16}/>
      </button>
    </>
  );
};

// ---- Stateful screen wrapper ------------------------------------------

const VoteScreen = ({ initialView = 'A', initialVotes = null, frozen = false }) => {
  const [view, setView] = React.useState(initialView);
  const [votes, setVotes] = React.useState(initialVotes || {});
  const [modal, setModal] = React.useState(initialView === 'B');

  const handleSubmit = (v) => { if (frozen) return; setVotes(v); setModal(true); };
  const handleConfirm = () => { if (frozen) return; setModal(false); setView('C'); };
  const handleCancel = () => { if (frozen) return; setModal(false); };

  // Frozen: derive directly from initial props
  const renderView = frozen ? initialView : view;
  const renderVotes = frozen ? (initialVotes || {}) : votes;
  const showModal = frozen ? (initialView === 'B') : modal;

  return (
    <div className={`vote-screen${(renderView === 'C' || renderView === 'D') ? ' is-resolved' : ''}`}>
      {(renderView === 'A' || renderView === 'B') && (
        <ViewA
          initialVotes={renderVotes}
          onSubmit={handleSubmit}
          frozen={frozen || renderView === 'B'}
        />
      )}
      {renderView === 'C' && <ViewC/>}
      {renderView === 'D' && <ViewD/>}
      {showModal && (
        <VoteModal
          votes={renderVotes}
          onConfirm={handleConfirm}
          onCancel={handleCancel}
        />
      )}
    </div>
  );
};

Object.assign(window, {
  VOTE_DATA, VHeader, TeamBanner, StarRow, TeammateRow,
  ViewA, ViewC, ViewD, VoteModal, VoteScreen,
});
