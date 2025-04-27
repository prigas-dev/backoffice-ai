function run({ username }) {
  query("UPDATE user SET username = ?;", username);

  return { username };
}
