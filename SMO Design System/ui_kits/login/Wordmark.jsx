// SMO — Wordmark component
// Off-white "SMO" with a blue dot beat after it. The dot is the only ornament.

const Wordmark = ({ size = 40 }) => {
  // Tuned dot/text proportions for crisp rendering at common sizes.
  const dot = Math.max(4, Math.round(size * 0.13));
  const lift = Math.round(size * 0.07);
  return (
    <span style={{ display: 'inline-flex', alignItems: 'baseline', userSelect: 'none' }} aria-label="SMO">
      <span style={{
        fontFamily: "'Inter', ui-sans-serif, system-ui, sans-serif",
        fontWeight: 700,
        fontSize: `${size}px`,
        letterSpacing: '-0.03em',
        color: '#F5F6F8',
        lineHeight: 1,
      }}>SMO</span>
      <span style={{
        display: 'inline-block',
        width: `${dot}px`,
        height: `${dot}px`,
        borderRadius: '999px',
        background: '#2080FF',
        marginLeft: `${Math.max(3, Math.round(size * 0.1))}px`,
        transform: `translateY(-${lift}px)`,
      }} />
    </span>
  );
};

Object.assign(window, { Wordmark });
