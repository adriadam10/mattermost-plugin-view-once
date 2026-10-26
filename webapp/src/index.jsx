import React from 'react';

import manifest from '../../plugin.json';

import {createBus} from './bus.js';
import Card from './card.jsx';
import Composer from './composer.jsx';
import {translator} from './i18n.js';
import {Icon} from './modal.jsx';
import {STYLES} from './styles.js';
import Viewer from './viewer.jsx';

const POST_TYPE = 'custom_view_once';

function currentLocale(state) {
    const {currentUserId, profiles} = state.entities.users;
    return profiles[currentUserId]?.locale || document.documentElement.lang || 'en';
}

// The composer and the viewer, opened from the editor button and the post cards.
function Modals({bus, t}) {
    const [compose, setCompose] = React.useState(null);
    const [view, setView] = React.useState(null);
    React.useEffect(() => {
        const offCompose = bus.on('compose', setCompose);
        const offView = bus.on('view', setView);
        return () => {
            offCompose();
            offView();
        };
    }, [bus]);
    const closeCompose = React.useCallback(() => setCompose(null), []);
    const closeView = React.useCallback(() => setView(null), []);
    return (
        <>
            {compose && (
                <Composer
                    request={compose}
                    onDone={closeCompose}
                    t={t}
                />
            )}
            {view && (
                <Viewer
                    key={view.id}
                    request={view}
                    onDone={closeView}
                    t={t}
                />
            )}
        </>
    );
}

// Rendered by Mattermost in the message box toolbar, with the box's draft.
function EditorButton({draft, updateText, bus, t}) {
    if (!draft?.channelId) {
        return null;
    }
    return (
        <button
            type='button'
            className='vo-editor-btn'
            title={t('button')}
            aria-label={t('button')}
            onClick={() => bus.emit('compose', {
                channelId: draft.channelId,
                rootId: draft.rootId,
                text: draft.message,
                clearText: () => updateText(''),
            })}
        >
            <Icon/>
        </button>
    );
}

class Plugin {
    initialize(registry, store) {
        const locale = currentLocale(store.getState());
        const t = translator(locale);
        const bus = createBus();

        this.style = document.createElement('style');
        this.style.textContent = STYLES;
        document.head.appendChild(this.style);

        registry.registerRootComponent(() => (
            <Modals
                bus={bus}
                t={t}
            />
        ));
        registry.registerPostEditorActionComponent((props) => (
            <EditorButton
                {...props}
                bus={bus}
                t={t}
            />
        ));
        registry.registerPostTypeComponent(POST_TYPE, (props) => (
            <Card
                post={props.post}
                bus={bus}
                locale={locale}
                t={t}
            />
        ));
        registry.registerWebSocketEventHandler(`custom_${manifest.id}_changed`, (msg) => bus.emit('changed', msg.data.id));
        registry.registerReconnectHandler(() => bus.emit('changed', null));
    }

    uninitialize() {
        this.style?.remove();
    }
}

window.registerPlugin(manifest.id, new Plugin());
