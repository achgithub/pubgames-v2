import React from 'react';

export const GameBoard = ({ board, pendingMove, onMove, disabled }) => {
  return (
    <div className="game-board">
      {board.map((cell, index) => {
        const isPending = pendingMove && pendingMove.position === index;
        const displayValue = isPending ? pendingMove.symbol : cell;
        const cellClass = isPending ? 'pending' : displayValue.toLowerCase();
        
        return (
          <div
            key={index}
            className={`game-cell ${displayValue ? 'filled' : ''} ${cellClass}`}
            onClick={() => !displayValue && !disabled && onMove(index)}
            onTouchEnd={(e) => {
              e.preventDefault();
              if (!displayValue && !disabled) onMove(index);
            }}
            style={{ cursor: disabled ? 'wait' : (displayValue ? 'not-allowed' : 'pointer') }}
          >
            {displayValue}
          </div>
        );
      })}
    </div>
  );
};
