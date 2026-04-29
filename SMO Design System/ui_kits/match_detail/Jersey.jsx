/* SMO — <Jersey>
 *
 * Stylised football jersey silhouette, drawn as inline SVG so it stays crisp
 * at any size and inherits color from CSS (`color: var(--team-red)` etc).
 *
 * Anatomy:
 *   - V-neck collar
 *   - Two short sleeves (slight outward angle)
 *   - Body that tapers in slightly toward the waist
 *   - Subtle inner highlight on the right side (10% white) for depth
 *
 * Usage:
 *   <Jersey color="red"  size={52} />
 *   <Jersey color="green" size={36} />
 *
 * The SVG uses `currentColor` for the fill, and the small highlight uses
 * white at 10% opacity on top of it. Keep it that way so the same component
 * works for both team colors via just a className swap on the parent.
 */
const Jersey = ({ color = 'red', size = 52, ariaLabel }) => {
  const cls = color === 'green' ? 'mc-jersey-green' : 'mc-jersey-red';
  return (
    <svg
      className={`mc-jersey ${cls}`}
      width={size}
      height={size}
      viewBox="0 0 64 64"
      xmlns="http://www.w3.org/2000/svg"
      role={ariaLabel ? 'img' : 'presentation'}
      aria-label={ariaLabel}
      aria-hidden={ariaLabel ? undefined : 'true'}
    >
      {/* Body — single closed path: shoulders → sleeves → waist → V-neck cutout */}
      <path
        d="
          M 22 6
          L 12 12
          L 6 22
          L 14 28
          L 18 24
          L 18 56
          C 18 57.66 19.34 59 21 59
          L 43 59
          C 44.66 59 46 57.66 46 56
          L 46 24
          L 50 28
          L 58 22
          L 52 12
          L 42 6
          L 36 12
          C 35 14.5 33.66 15.5 32 15.5
          C 30.34 15.5 29 14.5 28 12
          L 22 6
          Z
        "
        fill="currentColor"
      />
      {/* Subtle right-side highlight for a hint of depth — 10% white, very soft */}
      <path
        d="
          M 46 24
          L 46 56
          C 46 57.66 44.66 59 43 59
          L 36 59
          L 36 22
          L 46 24
          Z
        "
        fill="rgba(255,255,255,0.08)"
      />
      {/* V-neck inner accent (1px darker line to define collar) */}
      <path
        d="M 28 12 C 29 14.5 30.34 15.5 32 15.5 C 33.66 15.5 35 14.5 36 12"
        fill="none"
        stroke="rgba(0,0,0,0.25)"
        strokeWidth="1.2"
        strokeLinecap="round"
      />
    </svg>
  );
};

window.Jersey = Jersey;
