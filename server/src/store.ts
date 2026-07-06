// The rules of fakecloud, ported from the original Go store. Pure logic —
// no I/O — so the Durable Object stays a thin persistence shell around it.

export type Mode = "freeplay" | "duel";
export type Player = "X" | "O";

export interface Move {
  id: number;
  board_id: number;
  player: Player;
  position: number;
}

export interface Nameplate {
  id: number;
  board_id: number;
  text: string;
}

export interface Board {
  id: number;
  name: string;
  mode: Mode;
  cells: string[];
  next_player: string;
  winner: string;
  moves: Move[];
  nameplate?: Nameplate;
}

export interface State {
  nextBoardId: number;
  nextMoveId: number;
  nextNameplateId: number;
  boards: Record<number, { name: string; mode: Mode }>;
  moves: Record<number, Move>;
  nameplates: Record<number, Nameplate>;
}

export function emptyState(): State {
  return {
    nextBoardId: 1,
    nextMoveId: 1,
    nextNameplateId: 1,
    boards: {},
    moves: {},
    nameplates: {},
  };
}

// ApiError carries an HTTP status so the DO can map rule violations
// (409 cell taken, 404 not found, 400 bad input) straight to responses.
export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

const notFound = () => new ApiError(404, "not found");

const WINNING_LINES = [
  [0, 1, 2], [3, 4, 5], [6, 7, 8],
  [0, 3, 6], [1, 4, 7], [2, 5, 8],
  [0, 4, 8], [2, 4, 6],
];

/** Derives cells, next player, winner, moves, and nameplate for a board. */
function boardState(state: State, boardId: number): Board {
  const meta = state.boards[boardId];
  const board: Board = {
    id: boardId,
    name: meta.name,
    mode: meta.mode,
    cells: Array(9).fill(""),
    next_player: "",
    winner: "",
    moves: [],
  };
  for (const move of Object.values(state.moves)) {
    if (move.board_id === boardId) {
      board.cells[move.position] = move.player;
      board.moves.push(move);
    }
  }
  board.moves.sort((a, b) => a.id - b.id);

  for (const line of WINNING_LINES) {
    const [a, b, c] = line;
    if (board.cells[a] !== "" && board.cells[a] === board.cells[b] && board.cells[b] === board.cells[c]) {
      board.winner = board.cells[a];
      break;
    }
  }
  if (board.winner === "" && board.moves.length === 9) {
    board.winner = "draw";
  }

  // Turn order only means something in a refereed duel.
  if (meta.mode === "duel" && board.winner === "") {
    // X always starts; whoever has fewer marks on the board goes next.
    const countX = board.cells.filter((c) => c === "X").length;
    board.next_player = countX === board.moves.length - countX ? "X" : "O";
  }

  const plate = Object.values(state.nameplates).find((p) => p.board_id === boardId);
  if (plate) board.nameplate = plate;

  return board;
}

export function createBoard(state: State, name: unknown, mode: unknown): Board {
  if (typeof name !== "string" || name === "") {
    throw new ApiError(400, "name is required");
  }
  const resolved = mode === undefined || mode === "" ? "freeplay" : mode;
  if (resolved !== "freeplay" && resolved !== "duel") {
    throw new ApiError(400, 'mode must be "freeplay" or "duel"');
  }
  const id = state.nextBoardId++;
  state.boards[id] = { name, mode: resolved };
  return boardState(state, id);
}

export function listBoards(state: State): Board[] {
  return Object.keys(state.boards)
    .map(Number)
    .sort((a, b) => a - b)
    .map((id) => boardState(state, id));
}

export function getBoard(state: State, id: number): Board {
  if (!state.boards[id]) throw notFound();
  return boardState(state, id);
}

export function deleteBoard(state: State, id: number): void {
  if (!state.boards[id]) throw notFound();
  delete state.boards[id];
  for (const move of Object.values(state.moves)) {
    if (move.board_id === id) delete state.moves[move.id];
  }
  for (const plate of Object.values(state.nameplates)) {
    if (plate.board_id === id) delete state.nameplates[plate.id];
  }
}

export function createMove(state: State, boardId: unknown, player: unknown, position: unknown): Move {
  if (player !== "X" && player !== "O") {
    throw new ApiError(400, 'player must be "X" or "O"');
  }
  if (typeof position !== "number" || !Number.isInteger(position) || position < 0 || position > 8) {
    throw new ApiError(400, "position must be between 0 and 8");
  }
  if (typeof boardId !== "number" || !state.boards[boardId]) throw new ApiError(404, "board not found");

  const board = boardState(state, boardId);
  if (board.cells[position] !== "") {
    throw new ApiError(409, `position ${position} is already taken by ${board.cells[position]}`);
  }
  if (state.boards[boardId].mode === "duel") {
    if (board.winner !== "") {
      throw new ApiError(409, "game is already over");
    }
    if (player !== board.next_player) {
      throw new ApiError(409, `not ${player}'s turn, it is ${board.next_player}'s turn`);
    }
  }
  const move: Move = { id: state.nextMoveId++, board_id: boardId, player, position };
  state.moves[move.id] = move;
  return move;
}

export function getMove(state: State, id: number): Move {
  const move = state.moves[id];
  if (!move) throw notFound();
  return move;
}

export function deleteMove(state: State, id: number): void {
  if (!state.moves[id]) throw notFound();
  delete state.moves[id];
}

export function createNameplate(state: State, boardId: unknown, text: unknown): Nameplate {
  if (typeof text !== "string" || text === "") {
    throw new ApiError(400, "text is required");
  }
  if (typeof boardId !== "number" || !state.boards[boardId]) throw new ApiError(404, "board not found");
  const existing = Object.values(state.nameplates).find((p) => p.board_id === boardId);
  if (existing) {
    throw new ApiError(
      409,
      `board ${boardId} already has a nameplate (id=${existing.id}) — import it instead of creating another`,
    );
  }
  const plate: Nameplate = { id: state.nextNameplateId++, board_id: boardId, text };
  state.nameplates[plate.id] = plate;
  return plate;
}

export function getNameplate(state: State, id: number): Nameplate {
  const plate = state.nameplates[id];
  if (!plate) throw notFound();
  return plate;
}

export function updateNameplate(state: State, id: number, text: unknown): Nameplate {
  if (typeof text !== "string" || text === "") {
    throw new ApiError(400, "text is required");
  }
  const plate = state.nameplates[id];
  if (!plate) throw notFound();
  plate.text = text;
  return plate;
}

export function deleteNameplate(state: State, id: number): void {
  if (!state.nameplates[id]) throw notFound();
  delete state.nameplates[id];
}
