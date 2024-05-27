from sys import argv
from os import listdir, getcwd, mkdir
from os.path import join


if not len(argv) == 2:
    print("Requires exactly one argument")
    exit(1)

target = argv[1]

if target not in listdir(getcwd()):
    mkdir(target)

with open(join(getcwd(), target, f"{target}.go"), "w") as target_file:
    target_type = target.title()

    target_file.write(f"""
package {target}

import (
"sync"

"github.com/Cedi-Search/Cedi-Search-Engine/database"
)

type {target_type} struct {{
db *database.Database
}}

func New{target_type}(db *database.Database) *{target_type}{{
    return &{target_type}{{
        db: db,
    }}
}}

func ({target} *{target_type}) Index(wg *sync.WaitGroup) {{}}

func ({target} *{target_type}) Sniff(wg *sync.WaitGroup) {{}}

func ({target} *{target_type}) String() string {{return "{target_type}"}}

""")
