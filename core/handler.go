package core

// Handler is the callback signature used by widgets to react to events.
// Handlers are registered through Widget.On and invoked when a matching
// event is delivered via Widget.Dispatch.
//
// Parameters:
//   - source: The widget that originated the event. For bubbled events this
//     identifies the originating child, not the widget that happens to own
//     the handler.
//   - event:  The event being delivered (for example "click" or "focus").
//   - data:   Optional event-specific payload. The number and types of the
//     values are determined by the event definition and must be agreed
//     between dispatcher and handler.
//
// Return value:
//   - true  — the handler consumed the event; propagation should stop and
//     further handlers or parent containers should not see it.
//   - false — the handler observed the event but did not consume it;
//     dispatching continues to any remaining handlers or bubbling stages.
//
// Invocation order:
//
// When several handlers are registered on the same widget for the same
// event, they are invoked in reverse registration order (last-in,
// first-out). The most recently registered handler therefore gets the
// first look at the event and can short-circuit earlier ones by returning
// true. This makes it straightforward for late-added observers — for
// example, a dialog that wants to intercept keys while open — to override
// or filter the behaviour installed by a widget's constructor without
// having to unregister anything.
type Handler func(source Widget, event Event, data ...any) bool
