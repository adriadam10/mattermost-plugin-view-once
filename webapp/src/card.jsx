import React from 'react';
import {useSelector} from 'react-redux';

import {deleteMessage, fetchInfo, fetchUsers} from './api.js';
import {summary} from './i18n.js';
import {Icon} from './modal.jsx';

function displayName(user) {
    if (!user) {
        return '…';
    }
    const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ');
    return user.nickname || fullName || user.username;
}

const formatWhen = (ms, locale) => new Date(ms).toLocaleString(locale, {dateStyle: 'short', timeStyle: 'short'});

// Profiles from the store, plus the ones it has not loaded yet.
function useProfiles(ids) {
    const known = useSelector((state) => state.entities.users.profiles);
    const [fetched, setFetched] = React.useState({});
    const missing = ids.filter((id) => !known[id] && !fetched[id]);
    const missingKey = missing.join(',');
    React.useEffect(() => {
        if (!missingKey) {
            return;
        }
        fetchUsers(missingKey.split(',')).then((users) => {
            setFetched((prev) => Object.fromEntries([...Object.entries(prev), ...users.map((u) => [u.id, u])]));
        });
    }, [missingKey]);
    return (id) => known[id] || fetched[id];
}

function Viewers({viewers, locale}) {
    const profile = useProfiles(viewers.map((v) => v.user_id));
    return (
        <div className='vo-viewers'>
            {[...viewers].sort((a, b) => a.opened_at - b.opened_at).map((v) => (
                <div
                    key={v.user_id}
                    className='vo-viewer'
                >
                    <img
                        src={`${window.basename || ''}/api/v4/users/${v.user_id}/image?_=${profile(v.user_id)?.last_picture_update || 0}`}
                        alt=''
                    />
                    <span>{displayName(profile(v.user_id))}</span>
                    <span className='vo-viewer__when'>{formatWhen(v.opened_at, locale)}</span>
                </div>
            ))}
        </div>
    );
}

function SenderActions({id, info, locale, t}) {
    const [showViewers, setShowViewers] = React.useState(false);
    const viewers = info.viewers || [];
    const remove = () => {
        if (window.confirm(t('confirmDelete'))) {
            deleteMessage(id).catch(() => null);
        }
    };
    return (
        <>
            <span className='vo-card__meta'>
                {`${t('openedBy')} ${viewers.length} ${t('of')} ${info.recipient_count} · ${t('expires')} ${formatWhen(info.expires_at, locale)}`}
            </span>
            <div className='vo-card__actions'>
                <button
                    type='button'
                    className='vo-btn vo-btn--tertiary vo-btn--small'
                    onClick={() => setShowViewers(!showViewers)}
                >
                    {t('whoOpened')}
                </button>
                <button
                    type='button'
                    className='vo-btn vo-btn--danger vo-btn--small'
                    onClick={remove}
                >
                    {t('deleteForAll')}
                </button>
            </div>
            {showViewers && (viewers.length ? (
                <Viewers
                    viewers={viewers}
                    locale={locale}
                />
            ) : <span className='vo-card__meta'>{t('nobodyYet')}</span>)}
        </>
    );
}

// The post of a view-once message. The post itself has no content: the card
// asks the server what the current user may do with it.
export default function Card({post, bus, locale, t}) {
    const id = post.props?.view_once_id;
    const [info, setInfo] = React.useState(null);
    const [gone, setGone] = React.useState(false);

    React.useEffect(() => {
        let current = true;
        const load = () => fetchInfo(id).then((data) => current && setInfo(data)).catch((err) => {
            if (current && err.status === 404) {
                setGone(true);
            }
        });
        load();
        // null: the websocket reconnected and any message may have changed meanwhile.
        const off = bus.on('changed', (changedId) => (changedId === null || changedId === id) && load());
        return () => {
            current = false;
            off();
        };
    }, [id, bus]);

    if (gone || !id) {
        return (
            <div className='vo-card vo-card--done'>
                <div className='vo-card__head'>
                    <Icon/>
                    <span>{t('gone')}</span>
                </div>
            </div>
        );
    }
    if (!info) {
        return <div className='vo-card vo-card--done'><span className='vo-card__meta'>{t('loading')}</span></div>;
    }
    const done = !info.is_sender && info.opened;
    return (
        <div className={`vo-card${done ? ' vo-card--done' : ''}`}>
            <div className='vo-card__head'>
                <Icon/>
                <span className='vo-card__title'>{done ? t('opened') : t('cardTitle')}</span>
                <span className='vo-card__meta'>{summary(info, t)}</span>
            </div>
            {info.is_sender && (
                <SenderActions
                    id={id}
                    info={info}
                    locale={locale}
                    t={t}
                />
            )}
            {!info.is_sender && !info.opened && (
                <div className='vo-card__actions'>
                    <button
                        type='button'
                        className='vo-btn vo-btn--primary vo-btn--small'
                        onClick={() => bus.emit('view', {id})}
                    >
                        {t('open')}
                    </button>
                </div>
            )}
        </div>
    );
}
