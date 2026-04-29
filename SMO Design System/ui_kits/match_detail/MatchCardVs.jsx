/* SMO — <MatchCard>
 *
 * Compact horizontal card representing a match between Équipe rouge and
 * Équipe verte. 3-column layout: team A | center (VS or trophy) | team B.
 *
 * Props:
 *   status:      'upcoming' | 'live' | 'finished' | 'closed' | 'cancelled'
 *   dateLabel:   string  e.g. "Jeu 7 mai"     (sans-serif, mono numbers handled inline)
 *   timeLabel?:  string  e.g. "19h30"          (only for 'upcoming')
 *   winner?:     'red' | 'green'                (only for 'finished' / 'closed')
 *   votesLabel?: string e.g. "12 votes enregistrés"  (only for 'closed' per spec)
 *   onClick?:    handler — whole card is a button
 *
 * Anti-patterns (per design spec): no scores ("3-2"), no player avatars,
 * no sponsors on the jerseys, no vertical separator lines between columns,
 * no duplicated context badge in the footer.
 */

// --- Center pieces ---------------------------------------------------------

const CenterVS = () => <span className="mc-center-vs">VS</span>;

// Custom filled trophy SVG. Lucide's `trophy` is outline-only; we want a
// punchy filled mark for finished matches. Keeps the proportions of Lucide
// (pedestal + handles + bowl) so it sits naturally next to other Lucide icons
// elsewhere in the app.
const TrophyFilled = ({ color = 'red', size = 36, ariaLabel }) => {
  const cls = color === 'green' ? 'mc-center-trophy-green' : 'mc-center-trophy-red';
  return (
    <svg
      className={`mc-center-trophy ${cls}`}
      width={size}
      height={size}
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
      role={ariaLabel ? 'img' : 'presentation'}
      aria-label={ariaLabel}
      aria-hidden={ariaLabel ? undefined : 'true'}
    >
      {/* Side handles */}
      <path d="M5 4 L5 8 C5 9.5 6 10.7 7.4 11.1 L7.4 9 C6.6 8.7 6.2 8.1 6.2 7.4 L6.2 5.2 L7.5 5.2 L7.5 4 Z" fill="currentColor"/>
      <path d="M19 4 L19 8 C19 9.5 18 10.7 16.6 11.1 L16.6 9 C17.4 8.7 17.8 8.1 17.8 7.4 L17.8 5.2 L16.5 5.2 L16.5 4 Z" fill="currentColor"/>
      {/* Bowl */}
      <path d="M7 3 L17 3 L17 9 C17 11.76 14.76 14 12 14 C9.24 14 7 11.76 7 9 Z" fill="currentColor"/>
      {/* Stem */}
      <rect x="11" y="14" width="2" height="4" fill="currentColor"/>
      {/* Base */}
      <path d="M7 19 L17 19 L17 21 L7 21 Z" fill="currentColor"/>
      <rect x="9" y="17.5" width="6" height="2" rx="0.5" fill="currentColor"/>
    </svg>
  );
};

const CenterCancelled = () => (
  // Lucide `x-circle`, hand-shaped to match stroke conventions (2px round caps)
  <svg
    className="mc-center-cancelled"
    width="32" height="32" viewBox="0 0 24 24" fill="none"
    stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
    aria-hidden="true"
  >
    <circle cx="12" cy="12" r="10"/>
    <line x1="15" y1="9"  x2="9"  y2="15"/>
    <line x1="9"  y1="9"  x2="15" y2="15"/>
  </svg>
);

// --- Context badge ---------------------------------------------------------

// Renders date + state. State styling depends on status:
//   upcoming → plain muted text, time in mono
//   finished → plain muted text, "Terminé"
//   closed   → plain muted text, "Clôturé"
//   live     → blue pill with pulsing dot, "En cours"
//   cancelled→ subtle pill, strikethrough, "Annulé"
const ContextBadge = ({ status, dateLabel, timeLabel }) => {
  if (status === 'live') {
    return (
      <div className="mc-context">
        <span className="mc-pill mc-pill-live">
          <span className="mc-dot" aria-hidden="true"/>
          <span>{dateLabel}</span>
          <span className="mc-context-sep">·</span>
          <span>En cours</span>
        </span>
      </div>
    );
  }
  if (status === 'cancelled') {
    return (
      <div className="mc-context">
        <span className="mc-pill mc-pill-cancelled">
          <span>{dateLabel}</span>
          <span className="mc-context-sep">·</span>
          <span>Annulé</span>
        </span>
      </div>
    );
  }
  // Plain text variants
  let trailing;
  if (status === 'upcoming') trailing = <span className="num">{timeLabel}</span>;
  else if (status === 'finished') trailing = <span>Terminé</span>;
  else if (status === 'closed')   trailing = <span>Clôturé</span>;
  return (
    <div className="mc-context">
      <span>{dateLabel}</span>
      <span className="mc-context-sep">·</span>
      {trailing}
    </div>
  );
};

// --- Main component --------------------------------------------------------

const MatchCard = ({
  status = 'upcoming',
  dateLabel,
  timeLabel,
  winner,
  votesLabel,
  onClick,
}) => {
  const isCancelled = status === 'cancelled';
  let center;
  if (status === 'finished' || status === 'closed') {
    center = <TrophyFilled color={winner} ariaLabel={`Vainqueur : équipe ${winner === 'green' ? 'verte' : 'rouge'}`}/>;
  } else if (status === 'cancelled') {
    center = <CenterCancelled/>;
  } else {
    center = <CenterVS/>;
  }

  return (
    <button
      type="button"
      className={`mc${isCancelled ? ' is-cancelled' : ''}`}
      onClick={onClick}
      aria-label={`Match ${dateLabel}, ${
        status === 'upcoming' ? `à venir à ${timeLabel}` :
        status === 'live' ? 'en cours' :
        status === 'finished' ? `terminé, équipe ${winner === 'green' ? 'verte' : 'rouge'} gagnante` :
        status === 'closed' ? `clôturé, équipe ${winner === 'green' ? 'verte' : 'rouge'} gagnante` :
        'annulé'
      }`}
    >
      <ContextBadge status={status} dateLabel={dateLabel} timeLabel={timeLabel}/>

      <div className="mc-vs">
        <div className="mc-team">
          <Jersey color="red" size={52}/>
          <span className="mc-team-label">Équipe rouge</span>
        </div>

        <div className="mc-center">{center}</div>

        <div className="mc-team">
          <Jersey color="green" size={52}/>
          <span className="mc-team-label">Équipe verte</span>
        </div>
      </div>

      {status === 'closed' && votesLabel && (
        <div className="mc-foot">
          <span><span className="num">{votesLabel.match(/^\d+/)?.[0] || ''}</span>{votesLabel.replace(/^\d+/, '')}</span>
        </div>
      )}
    </button>
  );
};

window.MatchCardVs = MatchCard;
