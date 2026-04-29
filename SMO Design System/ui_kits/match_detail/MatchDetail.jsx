// SMO — Match Detail page composition
//
// Two modes:
//   - Mode 1 (edit): Match.Open. Organiser swaps players via drag-and-drop.
//     Bottom sticky "Valider les équipes" CTA is visible.
//   - Mode 2 (view): TeamsReady / InProgress / Completed / Closed. Read-only.
//     No CTA; jerseys + names + (optionally) score chips.
//
// Drag mechanics — Pointer Events (works mouse + touch + pen):
//   - pointerdown on a pion → start a drag, record offset between cursor and
//     pion center.
//   - pointermove on the field → translate the pion under the cursor; detect
//     hover over another pion via elementsFromPoint.
//   - pointerup → if released over another pion, swap their team membership.
//     Otherwise, snap back (the pion returns to its slot via CSS transition).
//
// Initial team rosters per spec.

const RED_INITIAL = [
  { id: 'al', initials: 'AL', name: 'Alex L.' },
  { id: 'ir', initials: 'IR', name: 'Inès R.' },
  { id: 'tb', initials: 'TB', name: 'Théo B.' },
  { id: 'mr', initials: 'MR', name: 'Marc R.' },
  { id: 'ps', initials: 'PS', name: 'Paul S.' },
];
const GREEN_INITIAL = [
  { id: 'ik', initials: 'IK', name: 'Issa K.' },
  { id: 'yn', initials: 'YN', name: 'Yann N.' },
  { id: 'cd', initials: 'CD', name: 'Cédric D.' },
  { id: 'sf', initials: 'SF', name: 'Sami F.' },
  { id: 'gp', initials: 'GP', name: 'Greg P.' },
];

// Sample bench (used to show the "Remplaçant" chip styling)
const BENCH_INITIAL = [
  { id: 'ld', initials: 'LD', name: 'Lola D.' },
  { id: 'mn', initials: 'MN', name: 'Marin N.' },
];

// Apply scores when in Mode 2 closed
function withScores(team, scores) {
  return team.map(p => ({ ...p, score: scores[p.id] ?? 4.0 }));
}
const SCORES = {
  al: 4.5, ir: 4.2, tb: 3.8, mr: 4.0, ps: 3.6,
  ik: 4.4, yn: 3.9, cd: 4.1, sf: 4.3, gp: 3.7,
};

// --- Composition Mode (drag-and-drop) ---

const CompositionMode = ({ initialRed, initialGreen }) => {
  const [red, setRed]     = React.useState(initialRed);
  const [green, setGreen] = React.useState(initialGreen);
  const [drag, setDrag]   = React.useState(null);
  // drag = { id, startX, startY, dx, dy, dropTargetId }

  const findPlayer = (id) => {
    const r = red.find(p => p.id === id);
    if (r) return { team: 'red', player: r };
    const g = green.find(p => p.id === id);
    if (g) return { team: 'green', player: g };
    return null;
  };

  // Swap two players' team memberships, preserving slot indices on each side.
  const swapTeams = (idA, idB) => {
    const a = findPlayer(idA), b = findPlayer(idB);
    if (!a || !b) return;
    if (a.team === b.team) {
      // Same team: reorder within the team — swap their slot positions.
      const team = a.team;
      const list = team === 'red' ? red : green;
      const idxA = list.findIndex(p => p.id === idA);
      const idxB = list.findIndex(p => p.id === idB);
      const next = list.slice();
      [next[idxA], next[idxB]] = [next[idxB], next[idxA]];
      (team === 'red' ? setRed : setGreen)(next);
    } else {
      // Cross-team: each goes to the other team, taking the other's slot index.
      const idxA = (a.team === 'red' ? red : green).findIndex(p => p.id === idA);
      const idxB = (b.team === 'red' ? red : green).findIndex(p => p.id === idB);
      const newRed = red.slice();
      const newGreen = green.slice();
      if (a.team === 'red') {
        newRed[idxA] = b.player;
        newGreen[idxB] = a.player;
      } else {
        newGreen[idxA] = b.player;
        newRed[idxB] = a.player;
      }
      setRed(newRed);
      setGreen(newGreen);
    }
  };

  const onPointerDown = (player, evt) => {
    evt.preventDefault();
    evt.target.setPointerCapture?.(evt.pointerId);
    const rect = evt.currentTarget.getBoundingClientRect();
    setDrag({
      id: player.id,
      pointerId: evt.pointerId,
      startX: evt.clientX,
      startY: evt.clientY,
      // pion center (anchored at translate -50% -50%, so center = rect center)
      pionCenterX: rect.left + rect.width / 2,
      pionCenterY: rect.top + rect.height / 2,
      dx: 0, dy: 0,
      dropTargetId: null,
    });

    const onMove = (e) => {
      const dx = e.clientX - rect.left - rect.width / 2;
      const dy = e.clientY - rect.top  - rect.height / 2;
      // Find pion under the cursor (skip the dragging one)
      const els = document.elementsFromPoint(e.clientX, e.clientY);
      let dropTargetId = null;
      for (const el of els) {
        const candidate = el.closest?.('.md-pion');
        if (candidate && candidate.dataset.pionId !== player.id) {
          dropTargetId = candidate.dataset.pionId;
          break;
        }
      }
      setDrag(d => d ? { ...d, dx, dy, dropTargetId } : null);
    };
    const onUp = () => {
      setDrag(curr => {
        if (curr?.dropTargetId) swapTeams(curr.id, curr.dropTargetId);
        return null;
      });
      window.removeEventListener('pointermove', onMove);
      window.removeEventListener('pointerup', onUp);
      window.removeEventListener('pointercancel', onUp);
    };
    window.addEventListener('pointermove', onMove);
    window.addEventListener('pointerup', onUp);
    window.addEventListener('pointercancel', onUp);
  };

  return (
    <Field
      redTeam={red}
      greenTeam={green}
      mode="edit"
      draggingId={drag?.id || null}
      dropTargetId={drag?.dropTargetId || null}
      dragOffset={drag ? { dx: drag.dx, dy: drag.dy } : null}
      onPointerDown={onPointerDown}
    />
  );
};

// --- Top-level page wrapper ---

const MatchDetailPage = ({
  variant,                 // 'composition' | 'composition-dragging' | 'finished' | 'closed'
  matchProps,              // props for <MatchCardVs/> in header
  showBottomBar = false,
  showScores = false,
  redTeam = RED_INITIAL,
  greenTeam = GREEN_INITIAL,
  bench = BENCH_INITIAL,
  staticDragId = null,
}) => {
  const interactive = variant === 'composition';
  return (
    <div className={`md-screen${showBottomBar ? ' has-bottombar' : ''}`}>
      {/* Header */}
      <div className="md-header">
        <button className="md-icon-btn" aria-label="Retour"><IconArrowLeft size={22}/></button>
        <div className="md-header-mid md-match-card">
          <MatchCardVs {...matchProps}/>
        </div>
        <button className="md-icon-btn" aria-label="Plus d'options"><IconMoreVert size={22}/></button>
      </div>

      {/* Field */}
      <div className="md-field-wrap">
        <div className="md-field-meta">
          <span>Composition</span>
          <span className="md-field-meta-state">
            {variant === 'composition' && 'Compose les équipes'}
            {variant === 'composition-dragging' && 'Glisse pour échanger'}
            {variant === 'finished' && 'Match terminé · lecture seule'}
            {variant === 'closed' && 'Clôturé · scores affichés'}
          </span>
        </div>
        {interactive ? (
          <CompositionMode initialRed={redTeam} initialGreen={greenTeam}/>
        ) : (
          <Field
            redTeam={showScores ? withScores(redTeam, SCORES) : redTeam}
            greenTeam={showScores ? withScores(greenTeam, SCORES) : greenTeam}
            mode="view"
            showScores={showScores}
            staticDragId={staticDragId}
            dropTargetId={staticDragId === 'tb' ? 'cd' : null}
          />
        )}
      </div>

      {/* Présents */}
      <PresentList
        red={redTeam}
        green={greenTeam}
        bench={bench}
        mode={interactive ? 'edit' : 'view'}
      />

      {/* Sticky bottom bar (Mode 1 only) */}
      {showBottomBar && (
        <div className="md-bottombar">
          <div className="md-bottombar-inner">
            <button className="md-bottombar-btn">
              Valider les équipes
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

window.MatchDetailPage = MatchDetailPage;
