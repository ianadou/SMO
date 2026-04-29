// SMO — Icons for Invite page (Lucide-shaped)
const InvIcon = ({ children, size = 20, className = '', ...rest }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24"
    fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
    className={className} {...rest}>{children}</svg>
);

const IconCalendar  = (p) => <InvIcon {...p}><rect x="3" y="4" width="18" height="18" rx="2" ry="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></InvIcon>;
const IconClock     = (p) => <InvIcon {...p}><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></InvIcon>;
const IconMapPin    = (p) => <InvIcon {...p}><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></InvIcon>;
const IconUsers     = (p) => <InvIcon {...p}><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></InvIcon>;
const IconCheck     = (p) => <InvIcon {...p}><polyline points="20 6 9 17 4 12"/></InvIcon>;
const IconX         = (p) => <InvIcon {...p}><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></InvIcon>;
// Outline variant for the primary CTA
const IconCheckCircle = (p) => <InvIcon {...p}><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></InvIcon>;
// Filled variants for the result hero (no stroke, fill currentColor)
const IconCheckCircleFilled = (p) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={p.size || 64} height={p.size || 64} viewBox="0 0 24 24" fill="currentColor" {...p}>
    <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm-1.4 14.6L5.8 11.8l1.4-1.4 3.4 3.4 6.2-6.2 1.4 1.4-7.6 7.6z" fillRule="evenodd"/>
  </svg>
);
const IconXCircleFilled = (p) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={p.size || 64} height={p.size || 64} viewBox="0 0 24 24" fill="currentColor" {...p}>
    <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm4.3 13l-1.3 1.3-3-3-3 3-1.3-1.3 3-3-3-3 1.3-1.3 3 3 3-3 1.3 1.3-3 3z" fillRule="evenodd"/>
  </svg>
);

Object.assign(window, {
  IconCalendar, IconClock, IconMapPin, IconUsers,
  IconCheck, IconX, IconCheckCircle,
  IconCheckCircleFilled, IconXCircleFilled,
});
