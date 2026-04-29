// SMO — GroupCard
// One row per group. Whole card is a button. Subtle hover lift via surface change.

const ResultDot = ({ color }) => (
  <span style={{
    display: 'inline-block',
    width: 8, height: 8,
    borderRadius: '999px',
    background: color,
    flexShrink: 0,
    marginRight: 8,
    transform: 'translateY(-1px)',
  }} />
);

const formatNextMatch = (m) => {
  if (!m) return null;
  return `Prochain match · ${m.day} · ${m.time}`;
};

const GroupCard = ({ group, onClick }) => {
  const { name, players, members, nextMatch, lastResult } = group;

  let resultNode;
  if (lastResult?.kind === 'red') {
    resultNode = <><ResultDot color="#DC2A3B" />Dernier match · Équipe rouge a gagné</>;
  } else if (lastResult?.kind === 'green') {
    resultNode = <><ResultDot color="#30D158" />Dernier match · Équipe verte a gagné</>;
  } else if (lastResult?.kind === 'live') {
    resultNode = <><ResultDot color="#2080FF" />Match en cours</>;
  } else {
    resultNode = <span className="muted">Pas encore de match terminé</span>;
  }

  return (
    <button type="button" className="g-card" onClick={onClick}>
      <div className="g-card-row g-card-head">
        <div className="g-card-title">{name}</div>
        <IconChevronRight size={18} className="g-card-chev" />
      </div>

      <div className="g-card-row g-card-players">
        <AvatarCluster members={members} size={28} />
        <span className="num g-card-count">{players}</span>
        <span className="g-card-count-label">joueurs</span>
      </div>

      <div className="g-card-row g-card-meta">
        {nextMatch ? (
          <>
            <div className="g-card-line">
              <span className="g-card-line-label">{formatNextMatch(nextMatch)}</span>
            </div>
            <div className="g-card-line g-card-venue">
              <IconMapPin size={13} />
              <span>{nextMatch.venue}</span>
            </div>
          </>
        ) : (
          <div className="g-card-line muted">Aucun match planifié</div>
        )}
      </div>

      <div className="g-card-row g-card-result">{resultNode}</div>
    </button>
  );
};

Object.assign(window, { GroupCard });
