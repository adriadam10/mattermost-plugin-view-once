import React from 'react';

// The view-once icon: a dashed circle with a 1, like WhatsApp's.
export function Icon({size = 18}) {
    return (
        <svg
            viewBox='0 0 24 24'
            width={size}
            height={size}
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            strokeLinejoin='round'
            aria-hidden='true'
        >
            <circle
                cx='12'
                cy='12'
                r='9'
                strokeDasharray='4.2 2.4'
            />
            <path d='M10.2 9.6 12.6 8v8'/>
        </svg>
    );
}

export function Modal({title, onClose, wide, footer, children, t}) {
    React.useEffect(() => {
        const onKey = (event) => event.key === 'Escape' && onClose();
        document.addEventListener('keydown', onKey);
        return () => document.removeEventListener('keydown', onKey);
    }, [onClose]);

    return (
        <div className='vo-backdrop'>
            <div
                className={`vo-modal${wide ? ' vo-modal--wide' : ''}`}
                role='dialog'
                aria-modal='true'
                aria-label={title}
            >
                <div className='vo-modal__header'>
                    <Icon size={20}/>
                    <h2 className='vo-modal__title'>{title}</h2>
                    <button
                        type='button'
                        className='vo-modal__close'
                        aria-label={t('close')}
                        onClick={onClose}
                    >
                        {'×'}
                    </button>
                </div>
                <div className='vo-modal__body'>{children}</div>
                {footer && <div className='vo-modal__footer'>{footer}</div>}
            </div>
        </div>
    );
}
