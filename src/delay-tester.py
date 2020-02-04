import pygame
import pygame.gfxdraw
import sys
import numpy as np

class delayTester(object):

    def __init__(self):
        # initialize pygame and graphics
        pygame.init()
        self.clock = pygame.time.Clock()
        self.FRAME_RATE = 60
        self.SCREEN_WIDTH, self.SCREEN_HEIGHT = 1920, 1080
        self.screen = pygame.display.set_mode(
            (self.SCREEN_WIDTH, self.SCREEN_HEIGHT), pygame.FULLSCREEN)
        self.BG_COLOR = 40,40,40
        self.TICK_COLOR = 150,150,150
        self.CURSOR_COLOR = 20,200,20

        # constants
        self.SCREEN_HEIGHT_IN_MS = 500
        self.CURSOR_VELOCITY = self.SCREEN_HEIGHT/float(self.SCREEN_HEIGHT_IN_MS/1000.) # pixels per second
        self.CURSOR_DELAY_SMALL_STEP = 1 # ms
        self.CURSOR_DELAY_BIG_STEP = 5 # ms
        self.NUM_TICKS = 7
        self.TICK_HEIGHT_IN_PX = 3
        self.TICK_WIDTH_IN_PX = .25*self.SCREEN_WIDTH
        self.CURSOR_HEIGHT_IN_PX = 20
        self.CURSOR_WIDTH_IN_PX = .75*self.TICK_WIDTH_IN_PX 

        # variables
        self.cursor_moving = True
        self.cursor_delay = 0 # ms
        self.cursor_ypos = self.SCREEN_HEIGHT

    def check_input(self):
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                self.quit()
            elif event.type == pygame.KEYDOWN:
                if event.key == pygame.K_SPACE:
                    self.cursor_moving = not(self.cursor_moving)
                if event.key == pygame.K_1:
                    self.cursor_delay += self.CURSOR_DELAY_BIG_STEP
                if event.key == pygame.K_2:
                    self.cursor_delay += self.CURSOR_DELAY_SMALL_STEP
                if event.key == pygame.K_3:
                    self.cursor_delay -= self.CURSOR_DELAY_SMALL_STEP
                if event.key == pygame.K_4:
                    self.cursor_delay -= self.CURSOR_DELAY_BIG_STEP
                elif event.key == pygame.K_ESCAPE: self.quit()
            elif event.type == pygame.KEYUP:
                pass

    def run(self):
        while True:
            time_passed = self.clock.tick_busy_loop(self.FRAME_RATE)/1000.
            self.check_input()
            self.draw_background()
            self.draw_all_ticks()
            self.update_cursor(time_passed)
            self.draw_cursor()
            self.draw_delay_msg()
            pygame.display.flip()

    def draw_delay_msg(self):
        draw_msg(self.screen, str(self.cursor_delay)+' ms', pos=(.5*self.SCREEN_WIDTH,.5*self.SCREEN_HEIGHT))

    def draw_all_ticks(self):
        for tick in range(1, self.NUM_TICKS-1):
            draw_tick(self.screen, self.TICK_WIDTH_IN_PX, self.TICK_HEIGHT_IN_PX,
                self.TICK_COLOR, self.SCREEN_WIDTH, tick/float(self.NUM_TICKS-1)*self.SCREEN_HEIGHT)

    def update_cursor(self, time_passed):
        self.cursor_ypos -= time_passed*self.CURSOR_VELOCITY*self.cursor_moving
        self.cursor_ypos = np.mod(self.cursor_ypos,self.SCREEN_HEIGHT)
        self.cursor_delay_ypos = self.cursor_ypos+self.SCREEN_HEIGHT*self.cursor_delay/float(self.SCREEN_HEIGHT_IN_MS)
        self.cursor_delay_ypos = np.mod(self.cursor_delay_ypos,self.SCREEN_HEIGHT)

    def draw_cursor(self):
        x_coords_lhs = [0, self.CURSOR_WIDTH_IN_PX, self.CURSOR_WIDTH_IN_PX]
        y_coords_lhs = [self.cursor_ypos, self.cursor_ypos+self.CURSOR_HEIGHT_IN_PX, self.cursor_ypos-self.CURSOR_HEIGHT_IN_PX]
        x_coords_rhs = [self.SCREEN_WIDTH, self.SCREEN_WIDTH-self.CURSOR_WIDTH_IN_PX, self.SCREEN_WIDTH-self.CURSOR_WIDTH_IN_PX]
        y_coords_rhs = [self.cursor_delay_ypos, self.cursor_delay_ypos+self.CURSOR_HEIGHT_IN_PX, self.cursor_delay_ypos-self.CURSOR_HEIGHT_IN_PX]
        draw_filled_aapoly(self.screen, list(zip(x_coords_lhs, y_coords_lhs)), self.CURSOR_COLOR)
        draw_filled_aapoly(self.screen, list(zip(x_coords_rhs, y_coords_rhs)), self.CURSOR_COLOR)

    def draw_background(self):
        self.screen.fill(self.BG_COLOR)

    def quit(self):
        sys.exit()

def draw_tick(screen, width, height, color, screen_width, y):
    draw_left_rect(screen, width, 0.5*height, color, 0, y)
    draw_right_rect(screen, width, 0.5*height, color, screen_width, y)

def draw_left_rect(screen, width, half_height, color, x, y):
    x_coords = [x+width, x+width, x, x]
    y_coords = [y-half_height, y+half_height, y+half_height, y-half_height]
    draw_filled_aapoly(screen, list(zip(x_coords, y_coords)), color)

def draw_right_rect(screen, width, half_height, color, x, y):
    x_coords = [x, x, x-width, x-width]
    y_coords = [y-half_height, y+half_height, y+half_height, y-half_height]
    draw_filled_aapoly(screen, list(zip(x_coords, y_coords)), color)

def draw_filled_aapoly(screen, coords, color):
    pygame.gfxdraw.filled_polygon(screen, coords, color)
    pygame.gfxdraw.aapolygon(screen, coords, color)

def draw_msg(screen, text, color=(255,255,255),
             loc='center', pos=(1024/2,768/2), size=50,
             font='freesansbold.ttf'):
    font = pygame.font.Font(font, size)
    text_surf, text_rect = make_text(text, font, color)
    if loc == 'center':
        text_rect.center = pos
    elif loc == 'left':
        text_rect.center = pos
        text_rect.left = pos[0]
    elif loc == 'right':
        text_rect.center = pos
        text_rect.right = pos[0]
    screen.blit(text_surf, text_rect)

def make_text(text, font, color):
    text_surface = font.render(text, True, color)
    return text_surface, text_surface.get_rect()

if __name__ == "__main__":
    game = delayTester()
    game.run()
