function run() {
  const user = query("SELECT username FROM user LIMIT 1;");
  if (user.length === 0) {
    throw new Error("user not found");
  }

  const username = user[0][0];

  return { username };
}
