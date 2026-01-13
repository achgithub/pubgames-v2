import React from 'react';

export const RematchModal = ({ 
  game, 
  user, 
  rematchRequest, 
  countdown,
  onRequestRematch, 
  onRespondToRematch,
  onGoToStandings 
}) => {
  const isWinner = game.winner_id === user.id;
  const isDraw = !game.winner_id;

  return (
    <div className="modal" style={{
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
      <div className="modal-content rematch-modal" style={{
        background: 'white',
        padding: '30px',
        borderRadius: '12px',
        maxWidth: '500px',
        width: '100%',
        maxHeight: '90vh',
        overflowY: 'auto',
        boxShadow: '0 10px 40px rgba(0,0,0,0.5)',
        textAlign: 'center'
      }}>
        <h2>Game Over!</h2>
        
        {game.winner_id ? (
          <p style={{fontSize: '24px', margin: '20px 0'}}>
            {isWinner ? (
              <span style={{color: '#27ae60'}}>🎉 You Won!</span>
            ) : (
              <span style={{color: '#e74c3c'}}>You Lost</span>
            )}
          </p>
        ) : (
          <p style={{fontSize: '24px', margin: '20px 0', color: '#7f8c8d'}}>It's a Draw!</p>
        )}

        {game.first_to > 1 && (
          <p style={{fontSize: '18px', margin: '15px 0'}}>
            Final Score: {game.player1_score} - {game.player2_score}
          </p>
        )}

        {!rematchRequest && (
          <div>
            <p style={{fontSize: '20px', margin: '20px 0'}}>Play again?</p>
            <div className="rematch-buttons">
              <button className="rematch-yes" onClick={onRequestRematch}>
                Yes
              </button>
              <button className="rematch-no" onClick={onGoToStandings}>
                No
              </button>
            </div>
          </div>
        )}

        {rematchRequest && rematchRequest.requester_id === user.id && rematchRequest.status === 'pending' && (
          <div className="rematch-waiting">
            <p style={{fontSize: '18px', marginBottom: '10px'}}>Waiting for opponent...</p>
            {countdown !== null && (
              <div>
                <div className="countdown">{countdown}</div>
                <p style={{fontSize: '14px', opacity: 0.8}}>seconds remaining</p>
              </div>
            )}
            {countdown === null && (
              <p style={{fontSize: '14px', opacity: 0.8}}>Opponent has 20 seconds to respond</p>
            )}
          </div>
        )}

        {rematchRequest && rematchRequest.opponent_id === user.id && rematchRequest.status === 'pending' && (
          <div>
            <p style={{fontSize: '20px', margin: '20px 0'}}>
              {game.player1_id === rematchRequest.requester_id ? game.player1_name : game.player2_name} wants a rematch!
            </p>
            <div className="rematch-buttons">
              <button className="rematch-yes" onClick={() => onRespondToRematch(true)}>
                Yes
              </button>
              <button className="rematch-no" onClick={() => onRespondToRematch(false)}>
                No
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
