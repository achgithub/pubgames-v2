import React from 'react';

export const StatsView = ({ user, playerStats, leaderboard, gameHistory }) => {
  return (
    <div>
      <h2>Statistics & Leaderboard</h2>

      {/* Player Stats */}
      {playerStats && (
        <div className="admin-section">
          <h3>Your Stats</h3>
          <div className="summary-stats">
            <div className="stat-card">
              <p>Games Played</p>
              <h3>{playerStats.games_played}</h3>
            </div>
            <div className="stat-card active">
              <p>Games Won</p>
              <h3>{playerStats.games_won}</h3>
            </div>
            <div className="stat-card eliminated">
              <p>Games Lost</p>
              <h3>{playerStats.games_lost}</h3>
            </div>
            <div className="stat-card">
              <p>Draws</p>
              <h3>{playerStats.games_draw}</h3>
            </div>
            <div className="stat-card">
              <p>Win Rate</p>
              <h3>{playerStats.win_rate.toFixed(1)}%</h3>
            </div>
          </div>
        </div>
      )}

      {/* Leaderboard */}
      <div className="admin-section">
        <h3>Leaderboard (Top 20)</h3>
        {leaderboard.length === 0 ? (
          <p className="info-text">No games played yet!</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Rank</th>
                <th>Player</th>
                <th>Played</th>
                <th>Won</th>
                <th>Lost</th>
                <th>Draw</th>
                <th>Win %</th>
              </tr>
            </thead>
            <tbody>
              {leaderboard.map((player, index) => (
                <tr key={player.user_id} className={player.user_id === user.id ? 'active' : ''}>
                  <td><strong>#{index + 1}</strong></td>
                  <td><strong>{player.user_name}</strong></td>
                  <td>{player.games_played}</td>
                  <td className="active-text">{player.games_won}</td>
                  <td className="eliminated-text">{player.games_lost}</td>
                  <td>{player.games_draw}</td>
                  <td><strong>{player.win_rate.toFixed(1)}%</strong></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Game History */}
      <div className="admin-section">
        <h3>Recent Games</h3>
        {gameHistory.length === 0 ? (
          <p className="info-text">No completed games yet!</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Opponent</th>
                <th>Mode</th>
                <th>Series</th>
                <th>Score</th>
                <th>Result</th>
              </tr>
            </thead>
            <tbody>
              {gameHistory.map(game => {
                const isPlayer1 = user.id === game.player1_id;
                const opponent = isPlayer1 ? game.player2_name : game.player1_name;
                const myScore = isPlayer1 ? game.player1_score : game.player2_score;
                const oppScore = isPlayer1 ? game.player2_score : game.player1_score;
                const won = game.winner_id === user.id;
                const draw = !game.winner_id;
                
                return (
                  <tr key={game.id}>
                    <td>{new Date(game.completed_at).toLocaleDateString()}</td>
                    <td>{opponent}</td>
                    <td>
                      <span className={`badge badge-${game.mode === 'timed' ? 'active' : 'draft'}`}>
                        {game.mode}
                      </span>
                    </td>
                    <td>First to {game.first_to}</td>
                    <td>{myScore} - {oppScore}</td>
                    <td>
                      {draw ? (
                        <span className="badge">Draw</span>
                      ) : won ? (
                        <span className="badge badge-correct">Won</span>
                      ) : (
                        <span className="badge badge-wrong">Lost</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
