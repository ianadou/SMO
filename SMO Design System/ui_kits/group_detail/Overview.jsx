// SMO — Pieces for Group Detail page

const Avatar = ({ initials, size = 40 }) => (
  <span style={{
    width: size, height: size, borderRadius: '999px',
    background: '#222831', color: '#F5F6F8',
    fontFamily: "'Inter', sans-serif",
    fontSize: Math.round(size * 0.36), fontWeight: 500,
    display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
    flexShrink: 0,
  }}>{initials}</span>
);

/**
 * AvatarRing — avatar wrapped in a circular progress ring.
 * `score` is on a 0..max scale (default 5). The ring sweeps clockwise from 12 o'clock.
 * The track sits at low opacity; the fill uses --action-primary.
 */
const AvatarRing = ({ initials, size = 48, score = 0, max = 5, stroke = 3 }) => {
  const pct = Math.max(0, Math.min(1, score / max));
  const radius = (size - stroke) / 2;
  const circumference = 2 * Math.PI * radius;
  const dash = circumference * pct;
  const innerSize = size - stroke * 2 - 4; // 2px gap between ring and avatar
  return (
    <span style={{ position: 'relative', width: size, height: size, flexShrink: 0, display: 'inline-flex', alignItems: 'center', justifyContent: 'center' }}>
      <svg width={size} height={size} style={{ position: 'absolute', inset: 0, transform: 'rotate(-90deg)' }} aria-hidden="true">
        <circle cx={size/2} cy={size/2} r={radius}
          fill="none" stroke="rgba(255,255,255,0.08)" strokeWidth={stroke}/>
        <circle cx={size/2} cy={size/2} r={radius}
          fill="none" stroke="var(--action-primary)" strokeWidth={stroke}
          strokeLinecap="round"
          strokeDasharray={`${dash} ${circumference - dash}`}/>
      </svg>
      <Avatar initials={initials} size={innerSize}/>
    </span>
  );
};

const Header = ({ groupName }) => (
  <header className="gd-header">
    <button className="gd-icon-btn" aria-label="Retour"><IconArrowLeft size={22} /></button>
    <h1 className="gd-h1">{groupName}</h1>
    <button className="gd-icon-btn" aria-label="Paramètres"><IconSettings size={22} /></button>
  </header>
);

const Tabs = ({ active = 'overview' }) => (
  <nav className="gd-tabs" role="tablist">
    <button role="tab" className={`gd-tab${active==='overview'?' is-active':''}`}>Vue d'ensemble</button>
    <button role="tab" className={`gd-tab${active==='players'?' is-active':''}`}>Joueurs</button>
    <button role="tab" className={`gd-tab${active==='matches'?' is-active':''}`}>Matchs</button>
  </nav>
);

// ---- Tab 1: Overview ----
// Prochain match: uses the shared <MatchCardVs> (jerseys + VS) component,
// followed by a venue + confirmation row that's specific to this surface.
const NextMatchBlock = () => (
  <section className="gd-block">
    <div className="gd-block-eyebrow">Prochain match</div>
    <MatchCardVs
      status="upcoming"
      dateLabel="Jeu 7 mai"
      timeLabel="19h30"
    />
    <div className="gd-next-meta">
      <div className="gd-next-venue">
        <IconMapPin2 size={14} />
        <span>Salle Pierre Mendès</span>
      </div>
      <div className="gd-next-conf">
        <span className="num">8</span><span className="muted">/10 confirmés</span>
      </div>
    </div>
  </section>
);

const StatsBlock = () => (
  <section className="gd-block">
    <div className="gd-block-eyebrow">Statistiques rapides</div>
    <div className="gd-stats">
      <div className="gd-stat">
        <IconUsers size={18} className="gd-stat-icon"/>
        <div><div className="num gd-stat-num">12</div><div className="gd-stat-label">joueurs</div></div>
      </div>
      <div className="gd-stat">
        <IconCalendarCheck size={18} className="gd-stat-icon"/>
        <div><div className="num gd-stat-num">47</div><div className="gd-stat-label">matchs joués</div></div>
      </div>
      <div className="gd-stat">
        <IconStar size={18} className="gd-stat-icon"/>
        <div><div className="num gd-stat-num">36</div><div className="gd-stat-label">matchs notés</div></div>
      </div>
    </div>
  </section>
);

const TopPlayer = ({ initials, name, score, matches }) => (
  <div className="gd-pcard">
    <AvatarRing initials={initials} size={44} score={score} />
    <div className="gd-pcard-mid">
      <div className="gd-pcard-name">{name}</div>
      <div className="gd-pcard-meta"><span className="num">{matches}</span><span className="muted"> matchs joués</span></div>
    </div>
    <div className="gd-pcard-score-num"><span className="num">{score.toFixed(1)}</span><span className="muted">/5</span></div>
  </div>
);

const TopThreeBlock = () => (
  <section className="gd-block">
    <div className="gd-block-eyebrow">Top 3 du groupe</div>
    <div className="gd-stack">
      <TopPlayer initials="AL" name="Alex L." score={4.7} matches={32}/>
      <TopPlayer initials="IR" name="Inès R." score={4.4} matches={28}/>
      <TopPlayer initials="IK" name="Issa K." score={4.2} matches={25}/>
    </div>
    <a href="#players" className="gd-link-row">
      Voir tous les joueurs
      <IconArrowRight size={14}/>
    </a>
  </section>
);

const OverviewTab = () => (
  <div className="gd-tab-body">
    <NextMatchBlock />
    <StatsBlock />
    <TopThreeBlock />
  </div>
);

Object.assign(window, { Avatar, AvatarRing, Header, Tabs, OverviewTab });
