#include <stdio.h>
#include <stdlib.h>
#include <ctype.h>
#include "common.h"

int
main(int argc, char *argv[])
{
  char team[TEAM_MAX]    = {0};
  char category[CAT_MAX] = {0};
  char points_str[11]    = {0};
  char answer[500]       = {0};
  long points            = 0;

  if (-1 == cgi_init(argv)) {
    return 0;
  }

  /* Read in team and answer */
  while (1) {
    size_t len;
    char   key[20];

    len = cgi_item(key, sizeof(key));
    if (0 == len) break;
    switch (key[0]) {
      case 't':
        cgi_item(team, sizeof(team));
        break;
      case 'c':
        cgi_item(category, sizeof(category));
        break;
      case 'p':
        cgi_item(points_str, sizeof(points_str));
        points = atol(points_str);
        break;
      case 'a':
        cgi_item(answer, sizeof(answer));
        break;
    }
  }

  /* Validate category name (prevent directory traversal) */
  {
    char *p;

    for (p = category; *p; p += 1) {
      if ((! isalnum(*p)) && ('-' != *p)) {
        cgi_page("Invalid category", "");
      }
    }
  }

  /* Check answer (also assures category exists) */
  {
    char needle[400];

    my_snprintf(needle, sizeof(needle), "%ld %s", points, answer);

    if (! anchored_search(package_path("%s/answers.txt", category), needle, 0)) {
      cgi_page("Wrong answer", "");
    }
  }

  {
    int ret = award_points(team, category, points, "P");
    if (ret < 0) {
        cgi_fail(ret);
    }
  }

  cgi_page("Points awarded",
           ("<p>%d points for %s.</p>\n"
            "<p><a href=\"/puzzles.html\">Back to puzzles</a></p>\n"
            "<!-- awarded %d -->"),
           points, team, points);

  return 0;
}
