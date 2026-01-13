import React, { useState } from 'react';

export const ChallengeModal = ({ opponent, onSend, onCancel }) => {
  const [selectedMode, setSelectedMode] = useState('normal');
  const [moveTimeLimit, setMoveTimeLimit] = useState(30);
  const [firstTo, setFirstTo] = useState(1);

  const handleSend = () => {
    onSend(opponent.user_id, selectedMode, moveTimeLimit, firstTo);
  };

  return (
    <div className="modal" onClick={onCancel} style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      background: 'rgba(0,0,0,0.75)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 9999,
      padding: '20px'
    }}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{
        background: 'white',
        padding: '30px',
        borderRadius: '12px',
        maxWidth: '500px',
        width: '100%',
        maxHeight: '90vh',
        overflowY: 'auto',
        boxShadow: '0 10px 40px rgba(0,0,0,0.5)'
      }}>
        <h2>Challenge {opponent.user_name}</h2>
        
        <div style={{margin: '20px 0'}}>
          <label style={{display: 'block', marginBottom: '10px'}}>
            <strong>Game Mode:</strong>
          </label>
          <select 
            value={selectedMode} 
            onChange={(e) => setSelectedMode(e.target.value)}
            style={{width: '100%', padding: '10px', fontSize: '16px', borderRadius: '4px', border: '2px solid #ddd'}}
          >
            <option value="normal">Normal (No time limit)</option>
            <option value="timed">Timed (Move timer)</option>
          </select>
        </div>

        {selectedMode === 'timed' && (
          <div style={{margin: '20px 0'}}>
            <label style={{display: 'block', marginBottom: '10px'}}>
              <strong>Time per move (seconds):</strong>
            </label>
            <input
              type="number"
              min="10"
              max="300"
              value={moveTimeLimit}
              onChange={(e) => setMoveTimeLimit(parseInt(e.target.value))}
              style={{width: '100%', padding: '10px', fontSize: '16px', borderRadius: '4px', border: '2px solid #ddd'}}
            />
          </div>
        )}

        <div style={{margin: '20px 0'}}>
          <label style={{display: 'block', marginBottom: '10px'}}>
            <strong>First to:</strong>
          </label>
          <select 
            value={firstTo} 
            onChange={(e) => setFirstTo(parseInt(e.target.value))}
            style={{width: '100%', padding: '10px', fontSize: '16px', borderRadius: '4px', border: '2px solid #ddd'}}
          >
            <option value="1">1 win (Single game)</option>
            <option value="2">2 wins (Best of 3)</option>
            <option value="3">3 wins (Best of 5)</option>
            <option value="5">5 wins (Best of 9)</option>
            <option value="10">10 wins (Best of 19)</option>
            <option value="20">20 wins (Best of 39)</option>
          </select>
        </div>

        <div style={{display: 'flex', gap: '10px', marginTop: '20px'}}>
          <button onClick={handleSend} className="cta-button" style={{flex: 1}}>
            Send Challenge
          </button>
          <button onClick={onCancel} className="back-btn" style={{flex: 1}}>
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
};
