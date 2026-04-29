// SMO — Icons additionally needed by the Register flow.
// Reuses Icon component conventions from ../login/Icons.jsx (loaded first).

const RegIcon = ({ children, size = 20, className = '', ...rest }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24"
    fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
    className={className} {...rest}>{children}</svg>
);

const IconMailCheck = (p) => (
  <RegIcon {...p}>
    <path d="M22 13V6a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h9"/>
    <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/>
    <path d="m16 19 2 2 4-4"/>
  </RegIcon>
);

const IconCheckCircleOutline = (p) => (
  <RegIcon {...p}>
    <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
    <polyline points="22 4 12 14.01 9 11.01"/>
  </RegIcon>
);

const IconXCircle = (p) => (
  <RegIcon {...p}>
    <circle cx="12" cy="12" r="10"/>
    <line x1="15" y1="9" x2="9" y2="15"/>
    <line x1="9" y1="9" x2="15" y2="15"/>
  </RegIcon>
);

// Filled hero variants for transition/result screens
const IconCheckCircleFilledReg = ({ size = 80, ...rest }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="currentColor" {...rest}>
    <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm-1.4 14.6L5.8 11.8l1.4-1.4 3.4 3.4 6.2-6.2 1.4 1.4-7.6 7.6z" fillRule="evenodd"/>
  </svg>
);

const IconXCircleFilledReg = ({ size = 80, ...rest }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" fill="currentColor" {...rest}>
    <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm4.3 13l-1.3 1.3-3-3-3 3-1.3-1.3 3-3-3-3 1.3-1.3 3 3 3-3 1.3 1.3-3 3z" fillRule="evenodd"/>
  </svg>
);

// Filled mail-check hero (large, action-primary)
const IconMailCheckFilled = ({ size = 80, ...rest }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width={size} height={size} viewBox="0 0 24 24" {...rest}>
    {/* Envelope body, filled */}
    <path d="M3 5h18a1 1 0 0 1 1 1v9.5a1 1 0 0 1-1 1h-7.5l-1 .5-1-.5H3a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1z" fill="currentColor" opacity="0.18"/>
    {/* Envelope outline + flap */}
    <path d="M3 5h18a1 1 0 0 1 1 1v9.5a1 1 0 0 1-1 1h-7.5"
      fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
    <path d="M2 6l9.4 6a1 1 0 0 0 1.2 0L22 6"
      fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"/>
    <path d="M3 5h18a1 1 0 0 1 1 1v9.5a1 1 0 0 1-1 1h-7.5"
      fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" style={{display:'none'}}/>
    {/* Check badge bottom-right */}
    <circle cx="18" cy="18" r="5" fill="currentColor"/>
    <polyline points="15.6 18 17.4 19.8 20.4 16.8" fill="none" stroke="#0E1014" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
);

Object.assign(window, {
  IconMailCheck, IconCheckCircleOutline, IconXCircle,
  IconCheckCircleFilledReg, IconXCircleFilledReg, IconMailCheckFilled,
});
