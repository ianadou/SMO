// SMO — Groups screen pieces: Header, PageTitle, FAB, EmptyState, SkeletonCard, GroupsList

const GroupsHeader = ({ onProfile }) => (
  <header className="g-header">
    <div className="g-header-mark">
      <Wordmark size={28} />
    </div>
    <button type="button" className="g-icon-btn" aria-label="Profil" onClick={onProfile}>
      <IconCircleUser size={24} />
    </button>
  </header>
);

const PageTitle = ({ count }) => (
  <div className="g-pagetitle">
    <h1 className="g-h1">Mes groupes</h1>
    <div className="g-sub">
      <span className="num">{count}</span> {count > 1 ? 'groupes actifs' : 'groupe actif'}
    </div>
  </div>
);

const Fab = ({ onClick, label = 'Créer un groupe' }) => (
  <button
    type="button"
    className="g-fab"
    aria-label={label}
    onClick={onClick}
  >
    <IconPlus size={24} />
  </button>
);

const EmptyState = ({ onCreate }) => (
  <div className="g-empty">
    <div className="g-empty-icon" aria-hidden="true">
      <IconUsersRound size={56} />
    </div>
    <div className="g-empty-title">Aucun groupe pour l'instant</div>
    <div className="g-empty-sub">Crée ton premier groupe pour organiser tes matchs.</div>
    <button type="button" className="smo-btn smo-btn-primary g-empty-cta" onClick={onCreate}>
      <IconPlus size={16} />
      <span>Créer mon premier groupe</span>
    </button>
  </div>
);

const SkeletonCard = () => (
  <div className="g-card g-card-skeleton" aria-hidden="true">
    <div className="g-card-row">
      <div className="sk sk-title" />
    </div>
    <div className="g-card-row" style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
      <div className="sk sk-cluster" />
      <div className="sk sk-line-sm" />
    </div>
    <div className="g-card-row">
      <div className="sk sk-line" />
    </div>
    <div className="g-card-row">
      <div className="sk sk-line-short" />
    </div>
  </div>
);

// Sort: groups with a planned match come first (soonest first).
const sortGroups = (groups) => {
  return [...groups].sort((a, b) => {
    const aHas = !!a.nextMatch;
    const bHas = !!b.nextMatch;
    if (aHas && !bHas) return -1;
    if (!aHas && bHas) return 1;
    if (aHas && bHas) return (a.nextMatch.sortKey ?? 0) - (b.nextMatch.sortKey ?? 0);
    return 0;
  });
};

Object.assign(window, {
  GroupsHeader, PageTitle, Fab, EmptyState, SkeletonCard, sortGroups,
});
