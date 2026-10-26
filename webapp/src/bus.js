// Lets the editor button and the post cards open the modals, which live in the
// one root component, and tells cards when their message changed.
export function createBus() {
    const listeners = new Map();
    return {
        on(event, listener) {
            if (!listeners.has(event)) {
                listeners.set(event, new Set());
            }
            listeners.get(event).add(listener);
            return () => listeners.get(event).delete(listener);
        },
        emit(event, payload) {
            listeners.get(event)?.forEach((listener) => listener(payload));
        },
    };
}
