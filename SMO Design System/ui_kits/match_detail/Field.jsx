// SMO — Field component
//
// Renders the football pitch with painted lines + absolutely-positioned pions.
// Coordinates are in PERCENT (0–100) inside the field rectangle, so the layout
// scales fluidly across mobile portrait and desktop landscape.
//
// Formation 2-2-1, vu depuis chaque côté:
//   - Goalkeeper (1):   on the goal line
//   - Defenders (2):    horizontally spread, in front of GK
//   - Forwards (2):     horizontally spread, near halfway line
//
// In portrait: red team occupies the top half (rows 0..50), green the bottom (50..100).
// In landscape: red on the left (cols 0..50), green on the right (50..100).

// --- Position math -------------------------------------------------------

// Build slot coordinates for a team given the orientation.
// `side` is 'red' (own half = top in portrait / left in landscape)
//          'green' (own half = bottom in portrait / right in landscape).
// Returns an array of 5 {x, y} percentage tuples in formation order: GK, DEF×2, FWD×2.
function buildSlots(side, orientation) {
  // Normalised "depth" axis: 0 = own goal line, 1 = halfway.
  // 3 rows: GK at 0.08, defenders at 0.30, forwards at 0.46
  const depthGK  = 0.08;
  const depthDEF = 0.30;
  const depthFWD = 0.46;

  // Spread on the cross-axis: defenders narrower than forwards looks more natural
  const defSpread = 0.30; // ±15% from center
  const fwdSpread = 0.40; // ±20% from center

  // Convert "depth" + "cross" into x/y percentages within the field.
  // In portrait (orientation === 'portrait'), depth runs vertically (Y).
  // In landscape, depth runs horizontally (X).
  const slots = [
    { depth: depthGK,  cross: 0 },                    // GK
    { depth: depthDEF, cross: -defSpread / 2 },       // Left-back
    { depth: depthDEF, cross:  defSpread / 2 },       // Right-back
    { depth: depthFWD, cross: -fwdSpread / 2 },       // Left-fwd
    { depth: depthFWD, cross:  fwdSpread / 2 },       // Right-fwd
  ];

  return slots.map(({ depth, cross }) => {
    if (orientation === 'portrait') {
      // Y axis is depth. Red has own goal at y=0, green at y=100.
      const y = side === 'red'
        ? depth * 50         // 0% .. 50% (top half)
        : 100 - depth * 50;  // 100% .. 50% (bottom half, mirrored)
      const x = 50 + cross * 100;
      return { x, y };
    } else {
      // Landscape: X axis is depth. Red has own goal at x=0, green at x=100.
      const x = side === 'red'
        ? depth * 50
        : 100 - depth * 50;
      const y = 50 + cross * 100;
      return { x, y };
    }
  });
}

// --- Field component -----------------------------------------------------

const Field = ({
  redTeam,        // array of player objects {id, initials, name}
  greenTeam,      // same shape
  mode = 'view',  // 'edit' | 'view'
  showScores = false,
  draggingId = null,
  dropTargetId = null,
  dragOffset = null,  // {dx, dy} in pixels for the dragging pion
  onPointerDown,      // (player, evt) => void
  staticDragId = null, // for "Vue 2 — drag in progress" static demo
}) => {
  const fieldRef = React.useRef(null);
  const [orientation, setOrientation] = React.useState('portrait');

  // Watch the field's actual rendered dimensions and pick orientation
  // from the rendered aspect ratio. This means the demo galleries
  // work even when the viewport itself is wide.
  React.useEffect(() => {
    const el = fieldRef.current;
    if (!el) return;
    const ro = new ResizeObserver(() => {
      const r = el.getBoundingClientRect();
      setOrientation(r.width > r.height ? 'landscape' : 'portrait');
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  const redSlots   = buildSlots('red',   orientation);
  const greenSlots = buildSlots('green', orientation);

  return (
    <div className="md-field" ref={fieldRef}>
      {/* Painted lines */}
      <div className="md-field-line is-pbox-top"/>
      <div className="md-field-line is-pbox-bottom"/>
      <div className="md-field-line is-pspot-top"/>
      <div className="md-field-line is-pspot-bottom"/>
      <div className="md-field-line is-halfway"/>
      <div className="md-field-line is-circle"/>
      <div className="md-field-line is-center-spot"/>

      {/* Pions */}
      {redTeam.slice(0, 5).map((p, i) => (
        <Pion
          key={p.id}
          player={p}
          team="red"
          x={redSlots[i].x}
          y={redSlots[i].y}
          mode={mode}
          showScore={showScores}
          isDragging={draggingId === p.id || staticDragId === p.id}
          isDropTarget={dropTargetId === p.id}
          dragOffset={draggingId === p.id ? dragOffset : null}
          staticDragOffset={staticDragId === p.id ? { dx: 12, dy: -18 } : null}
          onPointerDown={onPointerDown}
        />
      ))}
      {greenTeam.slice(0, 5).map((p, i) => (
        <Pion
          key={p.id}
          player={p}
          team="green"
          x={greenSlots[i].x}
          y={greenSlots[i].y}
          mode={mode}
          showScore={showScores}
          isDragging={draggingId === p.id || staticDragId === p.id}
          isDropTarget={dropTargetId === p.id}
          dragOffset={draggingId === p.id ? dragOffset : null}
          staticDragOffset={staticDragId === p.id ? { dx: 12, dy: -18 } : null}
          onPointerDown={onPointerDown}
        />
      ))}
    </div>
  );
};

window.Field = Field;
