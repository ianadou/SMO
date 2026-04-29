// SMO — Icons for Vote page (Lucide-shaped, hand-tuned)
const VIcon = ({ children, size = 20, className = '', ...rest }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24"
    fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
    className={className} {...rest}>{children}</svg>
);

// Star — outline (used for inactive / hover ghost)
const IconStarOutline = (p) => (
  <VIcon {...p}>
    <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
  </VIcon>
);

// Star — filled (used for active state). Same shape, fill currentColor, no stroke jitter.
const IconStarFilled = ({ size = 20, className = '', ...rest }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24"
    fill="currentColor" stroke="currentColor" strokeWidth="2" strokeLinejoin="round"
    className={className} {...rest}>
    <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
  </svg>
);

const IconAlertTriangle = (p) => (
  <VIcon {...p}>
    <path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
    <line x1="12" y1="9" x2="12" y2="13"/>
    <line x1="12" y1="17" x2="12.01" y2="17"/>
  </VIcon>
);

const IconCheck = (p) => <VIcon {...p}><polyline points="20 6 9 17 4 12"/></VIcon>;
const IconArrowRight = (p) => <VIcon {...p}><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></VIcon>;
const IconTrendingUp = (p) => <VIcon {...p}><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></VIcon>;
const IconTrendingDown = (p) => <VIcon {...p}><polyline points="23 18 13.5 8.5 8.5 13.5 1 6"/><polyline points="17 18 23 18 23 12"/></VIcon>;

// Filled check-circle (Vue C hero)
const IconCheckCircleFilled = ({ size = 64, ...rest }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="currentColor" {...rest}>
    <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm-1.4 14.6L5.8 11.8l1.4-1.4 3.4 3.4 6.2-6.2 1.4 1.4-7.6 7.6z" fillRule="evenodd"/>
  </svg>
);

Object.assign(window, {
  IconStarOutline, IconStarFilled, IconAlertTriangle,
  IconCheck, IconArrowRight, IconTrendingUp, IconTrendingDown,
  IconCheckCircleFilled,
});
