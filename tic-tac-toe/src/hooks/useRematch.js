import { useState, useEffect } from 'react';

export const useRematch = (rematchRequest, user, onExpired, onCheckStatus) => {
  const [countdown, setCountdown] = useState(null);

  useEffect(() => {
    if (!rematchRequest || !user) {
      setCountdown(null);
      return;
    }

    // If rematch was declined or expired, call onExpired immediately
    if (rematchRequest.status === 'declined' || rematchRequest.status === 'expired') {
      if (onExpired) {
        onExpired();
      }
      return;
    }

    // If rematch was accepted, no countdown needed
    if (rematchRequest.status === 'accepted') {
      setCountdown(null);
      return;
    }

    // If I'm the requester and status is pending - show countdown
    if (rematchRequest.requester_id === user.id && rematchRequest.status === 'pending') {
      if (rematchRequest.expires_at) {
        const expiresAt = new Date(rematchRequest.expires_at);
        
        const updateCountdown = () => {
          const now = new Date();
          const secondsLeft = Math.max(0, Math.floor((expiresAt - now) / 1000));
          
          if (secondsLeft > 60) {
            // Still in the 20s "waiting" period
            setCountdown(null);
          } else if (secondsLeft > 0) {
            // In the 60s countdown period
            setCountdown(secondsLeft);
          } else {
            // Countdown finished - check status ONCE
            setCountdown(null);
            if (onCheckStatus) {
              onCheckStatus(); // This will call checkRematchStatus once
            } else if (onExpired) {
              onExpired();
            }
          }
        };
        
        // Initial check
        updateCountdown();
        
        // Update every 1 second for smooth countdown display
        const interval = setInterval(updateCountdown, 1000);
        
        return () => clearInterval(interval);
      }
    }
  }, [rematchRequest, user, onExpired, onCheckStatus]);

  return countdown;
};
