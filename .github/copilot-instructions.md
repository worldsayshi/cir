# Composability Rules for Cir

## Core Principles

1. **Single Source of Truth**: Maintain all application state in `AppState`. Never store state in individual components that should be shared.

2. **State Changes Through `updateState` Only**: All state mutations should flow through the `updateState` function to ensure consistent state transitions and automatic persistence.

3. **Pure Rendering Components**: Components should be "dumb renderers" that transform data into UI without maintaining their own complex state.

4. **Consistent Component APIs**: All components should follow the same pattern:
   - Constructor that initializes the component
   - `Render` method that updates the component's view based on current state

5. **Unidirectional Data Flow**: State flows down through component hierarchy; events flow up to trigger state changes.

6. **Clear Component Boundaries**: Components should have well-defined responsibilities and minimal dependencies on other components.

7. **Event Handlers in Application Layer**: Keep component event handlers in the application layer where they can trigger state updates.

8. **Separation of UI and Logic**: UI components should handle display only; business logic belongs in the application layer.