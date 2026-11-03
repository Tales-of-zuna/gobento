'use client';
import { Component, type ErrorInfo, type ReactNode } from 'react';

/** Props of {@link MascotErrorBoundary}. */
export interface MascotErrorBoundaryProps {
  readonly children: ReactNode;
  /** Called once when the scene throws or its code fails to load. */
  readonly onError: () => void;
}

/**
 * Catches anything the 3D scene throws (including a failed chunk download)
 * and renders nothing, so the static still underneath stays and the visitor
 * never sees an error.
 */
export class MascotErrorBoundary extends Component<MascotErrorBoundaryProps, { failed: boolean }> {
  override state = { failed: false };

  static getDerivedStateFromError(): { failed: boolean } {
    return { failed: true };
  }

  override componentDidCatch(error: unknown, info: ErrorInfo): void {
    if (process.env.NODE_ENV !== 'production') {
      console.warn('Mascot scene failed; showing the static still.', error, info.componentStack);
    }
    this.props.onError();
  }

  override render(): ReactNode {
    return this.state.failed ? null : this.props.children;
  }
}
