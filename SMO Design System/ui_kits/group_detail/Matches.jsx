// SMO — Matches tab
// Uses the shared <MatchCardVs> component (jerseys + VS / trophy layout).

const STATUS_PROPS = {
  upcoming:  { status: 'upcoming' },
  live:      { status: 'live' },
  finished:  { status: 'finished' },
  closed:    { status: 'closed' },
  cancelled: { status: 'cancelled' },
};

// Domain data for the Matches tab. `time` is only used by upcoming entries.
const MATCHES = [
  { date:'Jeu 30 avril', venue:'Salle Pierre Mendès', status:'finished', winner:'red'   },
  { date:'Jeu 24 avril', venue:'Salle Pierre Mendès', status:'finished', winner:'green' },
  { date:'Jeu 17 avril', venue:'Stade municipal',      status:'finished', winner:'red'   },
  { date:'Jeu 10 avril', venue:'Salle Pierre Mendès', status:'closed',   winner:'green', votesLabel:'12 votes enregistrés' },
  { date:'Jeu 3 avril',  venue:'Salle Pierre Mendès', status:'closed',   winner:'red',   votesLabel:'10 votes enregistrés' },
];

const MatchesTab = ({ empty = false }) => (
  <div className="gd-tab-body">
    <button className="smo-btn smo-btn-primary gd-create-match">
      <IconPlus2 size={16}/><span>Créer un match</span>
    </button>
    {empty ? (
      <div className="gd-empty-inline">
        <div className="gd-empty-icon"><IconCalendarX size={48}/></div>
        <div className="gd-empty-title">Aucun match pour ce groupe</div>
        <div className="gd-empty-sub">Crée le premier match pour commencer.</div>
      </div>
    ) : (
      <div className="gd-stack gd-stack-tight">
        {MATCHES.map((m, i) => (
          <MatchCardVs
            key={i}
            status={m.status}
            dateLabel={m.date}
            timeLabel={m.time}
            winner={m.winner}
            votesLabel={m.votesLabel}
          />
        ))}
      </div>
    )}
  </div>
);

Object.assign(window, { MatchesTab, MATCHES });
