// SMO — Players tab

const PlayerRow = ({ initials, name, score, matches }) => (
  <div className="gd-pcard gd-prow">
    <AvatarRing initials={initials} size={48} score={score} />
    <div className="gd-pcard-mid">
      <div className="gd-pcard-name">{name}</div>
      <div className="gd-pcard-meta"><span className="num">{matches}</span><span className="muted"> matchs joués</span></div>
    </div>
    <div className="gd-pcard-score-num"><span className="num">{score.toFixed(1)}</span><span className="muted">/5</span></div>
    <button className="gd-icon-btn gd-prow-more" aria-label="Actions"><IconMoreVert size={18}/></button>
  </div>
);

const AddPlayerForm = ({ error = '' }) => (
  <div className="gd-add">
    <div className="gd-add-label">Prénom ou pseudonyme</div>
    <div className="gd-add-row">
      <div className="smo-input-wrap gd-add-input-wrap">
        <input className="smo-input" placeholder="Ajouter un joueur…" />
      </div>
      <button className="smo-btn smo-btn-primary gd-add-btn">Ajouter</button>
    </div>
    {error && (
      <div className="smo-error" role="alert" style={{marginTop: 12}}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        <span>{error}</span>
      </div>
    )}
  </div>
);

const PLAYERS = [
  { initials:'AL', name:'Alex L.', score:4.7, matches:32 },
  { initials:'IR', name:'Inès R.', score:4.4, matches:28 },
  { initials:'IK', name:'Issa K.', score:4.2, matches:25 },
  { initials:'TB', name:'Théo B.', score:4.0, matches:30 },
  { initials:'MR', name:'Marc R.', score:3.9, matches:22 },
  { initials:'PS', name:'Paul S.', score:3.7, matches:18 },
  { initials:'YN', name:'Yann N.', score:3.5, matches:20 },
  { initials:'CD', name:'Cédric D.', score:3.4, matches:15 },
  { initials:'SF', name:'Sami F.', score:3.2, matches:14 },
  { initials:'GP', name:'Greg P.', score:3.0, matches:12 },
  { initials:'NL', name:'Nico L.', score:2.8, matches:10 },
  { initials:'OB', name:'Omar B.', score:2.5, matches:8 },
];

const PlayersTab = () => (
  <div className="gd-tab-body">
    <AddPlayerForm />
    <div className="gd-stack gd-stack-tight">
      {PLAYERS.map((p, i) => <PlayerRow key={i} {...p}/>)}
    </div>
  </div>
);

Object.assign(window, { PlayersTab, PLAYERS });
