// SMO — Pion (player token on the field)
//
// Visual: colored disc with initials, name pill underneath, optional score chip
// (Mode 2 closed only). Position is set via inline `left` / `top` percentages
// — coordinates anchor at the pion's center via translate(-50%, -50%).
//
// Drag mechanics live in the parent (MatchDetail.jsx). This component just
// receives `isDragging` / `isDropTarget` flags and renders accordingly, plus
// applies a pixel-offset transform during an active drag.
//
// Truncation rule for label: max 9 chars (so "Cédric D." passes), suffix "…"
// otherwise. Initials inside the disc are always 2 letters.

const truncate = (s, max = 9) => (s.length > max ? s.slice(0, max - 1) + '…' : s);

const Pion = ({
  player,
  team,
  x, y,                      // percentage within field
  mode,                      // 'edit' | 'view'
  showScore = false,
  isDragging = false,
  isDropTarget = false,
  dragOffset = null,         // {dx, dy} pixels — applied during active drag
  staticDragOffset = null,   // for the Vue 2 static demo
  onPointerDown,
}) => {
  const interactive = mode === 'edit' && !isDragging;

  // While dragging, override the smooth slot transition with a pixel-offset
  // transform so the pion glides under the cursor 1:1.
  const off = dragOffset || staticDragOffset;
  const style = {
    left: `${x}%`,
    top:  `${y}%`,
  };
  if (off) {
    style.transform = `translate(calc(-50% + ${off.dx}px), calc(-50% + ${off.dy}px))`;
  }

  const cls = [
    'md-pion',
    team === 'green' ? 'md-pion-green' : 'md-pion-red',
    interactive ? 'is-interactive' : '',
    isDragging ? 'is-dragging' : '',
    isDropTarget ? 'is-drop-target' : '',
  ].filter(Boolean).join(' ');

  const handlePointerDown = (e) => {
    if (mode !== 'edit') return;
    onPointerDown?.(player, e);
  };

  return (
    <div
      className={cls}
      style={style}
      data-pion-id={player.id}
      onPointerDown={handlePointerDown}
    >
      <div className="md-pion-disc">{player.initials}</div>
      <div className="md-pion-name">{truncate(player.name)}</div>
      {showScore && player.score != null && (
        <div className="md-pion-score">
          <span className="num">{player.score.toFixed(1)}</span>
        </div>
      )}
    </div>
  );
};

window.Pion = Pion;
