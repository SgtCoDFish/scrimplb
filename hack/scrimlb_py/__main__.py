import sys


def main(command: str):
    if command == "upload_backend_details":
        pass
    elif command == "fetch_backend":
        pass
    else:
        print("Invalid command", command, file=sys.stderr)


if __name__ == '__main__':
    args = sys.argv[1:]

    main(args[0])
