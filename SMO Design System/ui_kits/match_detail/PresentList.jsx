// SMO — Présents section
//
// Compact wrapping grid of chips listing every player who confirmed attendance
// for the match. Each chip shows initials avatar + truncated name. In edit mode,
// chips are clickable as an alternative to drag-and-drop (tap a chip then tap
// a pion to swap — same selection model as field tap-to-swap, but kept
// out-of-scope for the demo: chips just show :hover state).
//
// "Bench" players (present but not in the starting 10) are dimmed and tagged.

const PresentList = ({ red, green, bench = [], mode = 'view' }) => {
  const total = red.length + green.length + bench.length;
  const placed = red.length + green.length;
  return (
    <section className="md-presents">
      <div className="md-presents-head">
        <span className="md-presents-title">Présents</span>
        <span className="md-presents-count">
          <span className="num">{placed}</span>
          <span style={{ color: 'var(--fg-subtle)' }}> / </span>
          <span className="num" style={{ color: 'var(--fg-subtle)' }}>{total}</span>
        </span>
      </div>
      <div className="md-chips">
        {red.map(p => (
          <span key={p.id} className={`md-chip is-red${mode==='edit'?' is-actionable':''}`}>
            <span className="md-chip-avatar">{p.initials}</span>
            <span>{p.name}</span>
          </span>
        ))}
        {green.map(p => (
          <span key={p.id} className={`md-chip is-green${mode==='edit'?' is-actionable':''}`}>
            <span className="md-chip-avatar">{p.initials}</span>
            <span>{p.name}</span>
          </span>
        ))}
        {bench.map(p => (
          <span key={p.id} className="md-chip is-bench">
            <span className="md-chip-avatar">{p.initials}</span>
            <span>{p.name}</span>
            <span className="md-chip-bench-tag">· Remplaçant</span>
          </span>
        ))}
      </div>
    </section>
  );
};

window.PresentList = PresentList;
