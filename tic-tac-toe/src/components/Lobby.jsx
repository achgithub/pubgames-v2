import React from 'react';

export const Lobby = ({ 
  onlineUsers, 
  pendingChallenges, 
  onChallenge, 
  onRespondToChallenge,
  onRefresh
}) => {
  return (
    <div>
      <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px'}}>
        <h2 style={{margin: 0}}>Game Lobby</h2>
        <button onClick={onRefresh} className="btn-info" style={{padding: '8px 16px'}}>
          🔄 Refresh
        </button>
      </div>

      {/* Pending Challenges */}
      {pendingChallenges.length > 0 && (
        <div className="admin-section">
          <h3>Pending Challenges ({pendingChallenges.length})</h3>
          {pendingChallenges.map(challenge => (
            <div key={challenge.id} className="user-card">
              <div>
                <strong>{challenge.player1_name}</strong> challenged you to a {challenge.mode} game
                {challenge.mode === 'timed' && ` (${challenge.move_time_limit}s per move)`}
                {challenge.first_to > 1 && ` - First to ${challenge.first_to}`}
              </div>
              <div>
                <button 
                  onClick={() => onRespondToChallenge(challenge.id, true)} 
                  className="btn-success" 
                  style={{marginRight: '10px'}}
                >
                  Accept
                </button>
                <button 
                  onClick={() => onRespondToChallenge(challenge.id, false)} 
                  className="btn-warning"
                >
                  Decline
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Online Users */}
      <div className="admin-section">
        <h3>Online Players ({onlineUsers.length})</h3>
        {onlineUsers.length === 0 ? (
          <p className="info-text">No other players online right now. Check back soon!</p>
        ) : (
          <div>
            {onlineUsers.map(ou => (
              <div key={ou.user_id} className={`user-card ${ou.in_game ? 'in-game' : ''}`}>
                <div>
                  <strong>{ou.user_name}</strong>
                  {ou.in_game && <span className="badge badge-active" style={{marginLeft: '10px'}}>In Game</span>}
                </div>
                {!ou.in_game && (
                  <button onClick={() => onChallenge(ou)} className="btn-info">
                    Challenge
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="rules">
        <h3>How to Play</h3>
        <ul>
          <li>Challenge another online player to a game</li>
          <li>Choose <strong>Normal</strong> mode (no time limit) or <strong>Timed</strong> mode (move timer)</li>
          <li>Select <strong>First to X</strong> wins (1, 2, 3, 5, 10, or 20 wins)</li>
          <li>First player is X, second player is O</li>
          <li>Get three in a row (horizontal, vertical, or diagonal) to win a round!</li>
          <li>First to reach the target score wins the series</li>
          <li>After the game, you can request a rematch</li>
        </ul>
      </div>
    </div>
  );
};
