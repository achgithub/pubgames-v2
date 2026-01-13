import React from 'react';
import { GameBoard } from './GameBoard';

export const GameView = ({ game, user, pendingMove, onMove }) => {
  const isPlayer1 = user.id === game.player1_id;
  const isPlayer2 = game.player2_id && user.id === game.player2_id;
  const playerNumber = isPlayer1 ? 1 : 2;
  const isMyTurn = game.current_turn === playerNumber;

  return (
    <div>
      <h2>Game in Progress</h2>

      {/* Series Score (if first_to > 1) */}
      {game.first_to > 1 && (
        <div className="series-score">
          <h3>First to {game.first_to} Wins</h3>
          <div className="scores">
            <div>
              <div>{game.player1_name}</div>
              <div style={{fontSize: '36px'}}>{game.player1_score}</div>
            </div>
            <div style={{fontSize: '28px'}}>-</div>
            <div>
              <div>{game.player2_name}</div>
              <div style={{fontSize: '36px'}}>{game.player2_score}</div>
            </div>
          </div>
          <div className="round-info">Round {game.current_round}</div>
        </div>
      )}

      <div className="game-info-panel">
        <div style={{display: 'flex', justifyContent: 'space-between', marginBottom: '15px'}}>
          <div>
            <strong>{game.player1_name}</strong> (X)
            {game.current_turn === 1 && isPlayer1 && (
              <span className="badge badge-active" style={{marginLeft: '10px'}}>Your Turn</span>
            )}
          </div>
          <div>VS</div>
          <div>
            <strong>{game.player2_name}</strong> (O)
            {game.current_turn === 2 && isPlayer2 && (
              <span className="badge badge-active" style={{marginLeft: '10px'}}>Your Turn</span>
            )}
          </div>
        </div>

        {game.mode === 'timed' && (
          <div style={{textAlign: 'center', color: '#7f8c8d'}}>
            <small>Move Time Limit: {game.move_time_limit} seconds</small>
          </div>
        )}
      </div>

      <GameBoard 
        board={JSON.parse(game.board)}
        pendingMove={pendingMove}
        onMove={onMove}
        disabled={!!pendingMove}
      />

      <div style={{textAlign: 'center', marginTop: '20px'}}>
        {isMyTurn ? (
          <p style={{fontSize: '18px', color: isPlayer1 ? '#e74c3c' : '#3498db', fontWeight: 'bold'}}>
            Your turn! Place your {isPlayer1 ? 'X' : 'O'}
          </p>
        ) : (
          <p style={{fontSize: '18px', color: '#7f8c8d'}}>Waiting for opponent...</p>
        )}
      </div>
    </div>
  );
};
