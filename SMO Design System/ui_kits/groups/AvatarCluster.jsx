// SMO — AvatarCluster
// Stacked initials pills. Max 4 visible, then "+N" pill. No per-player color.

const Avatar = ({ initials, size = 28, z = 0, offset = 0 }) => (
  <span
    style={{
      width: size, height: size,
      borderRadius: '999px',
      background: '#222831',           // border-default — neutral, in palette
      color: '#F5F6F8',
      fontFamily: "'Inter', sans-serif",
      fontSize: Math.round(size * 0.4),
      fontWeight: 500,
      letterSpacing: '0.01em',
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      // Ring matches the card surface so cluster reads as overlapping.
      boxShadow: '0 0 0 2px #1A1F26',
      marginLeft: offset,
      position: 'relative',
      zIndex: z,
      flexShrink: 0,
    }}
  >{initials}</span>
);

const PlusN = ({ n, size = 28, z = 0, offset = 0 }) => (
  <span
    style={{
      width: size, height: size,
      borderRadius: '999px',
      background: '#0E1014',           // bg-base — sits darker than card
      color: '#F5F6F8',
      fontFamily: "'JetBrains Mono', monospace",
      fontSize: Math.round(size * 0.38),
      fontWeight: 500,
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      boxShadow: '0 0 0 2px #1A1F26',
      marginLeft: offset,
      position: 'relative',
      zIndex: z,
      flexShrink: 0,
    }}
  >+{n}</span>
);

const AvatarCluster = ({ members = [], max = 4, size = 28 }) => {
  const visible = members.slice(0, max);
  const overflow = Math.max(0, members.length - max);
  const overlap = -Math.round(size * 0.32);

  return (
    <span style={{ display: 'inline-flex', alignItems: 'center' }} aria-hidden="true">
      {visible.map((m, i) => (
        <Avatar
          key={i}
          initials={m}
          size={size}
          z={visible.length - i}
          offset={i === 0 ? 0 : overlap}
        />
      ))}
      {overflow > 0 && (
        <PlusN n={overflow} size={size} z={0} offset={overlap} />
      )}
    </span>
  );
};

Object.assign(window, { AvatarCluster });
